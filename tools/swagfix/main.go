// Command swagfix post-processes the OpenAPI spec emitted by swag v2 to work
// around two swag-v2 limitations:
//
//  1. https://github.com/swaggo/swag/issues/2086: swag wraps every request body
//     in a malformed `oneOf` whose first member is an empty `{type: object}`.
//     Clients like Postman generate their example body from that first member, so
//     the body comes out empty (and, with no body, they drop the Content-Type
//     header). We collapse the oneOf back to the real `$ref` it wraps so the
//     per-field `example` values are used.
//
//  2. swag annotations can only describe a JWT bearer token as a raw `apiKey`
//     header (swag supports apiKey/basic/oauth2 only, not `type: http`). Postman
//     then imports it as "API Key" auth and puts the bare variable into the
//     Authorization header without the required "Bearer " prefix, so every
//     protected request 401s. We rewrite that scheme into a proper HTTP Bearer
//     scheme, so Postman imports "Bearer Token" auth (and adds the prefix itself)
//     and Swagger UI's Authorize dialog asks only for the token.
//
// It rewrites <dir>/swagger.json and <dir>/swagger.yaml (dir defaults to
// "docs"). docs/docs.go is intentionally left untouched: the service serves the
// embedded swagger.json directly and bypasses the swag registry (see
// docs/embed.go), so the template there is unused at runtime.
//
// Both fixes are idempotent — once applied, re-running is a no-op.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

func main() {
	dir := "docs"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := run(dir); err != nil {
		fmt.Fprintln(os.Stderr, "swagfix:", err)
		os.Exit(1)
	}
}

func run(dir string) error {
	jsonPath := filepath.Join(dir, "swagger.json")

	//nolint:gosec // path is a build-time tool argument, not user input
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", jsonPath, err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse %s: %w", jsonPath, err)
	}

	fixed := fixRequestBodies(doc)
	secFixed := fixSecuritySchemes(doc)

	out, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	out = append(out, '\n')
	//nolint:gosec // build-time tool writing back its own generated spec
	if err := os.WriteFile(jsonPath, out, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", jsonPath, err)
	}

	yml, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	yamlPath := filepath.Join(dir, "swagger.yaml")
	//nolint:gosec // build-time tool writing back its own generated spec
	if err := os.WriteFile(yamlPath, yml, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", yamlPath, err)
	}

	fmt.Printf("swagfix: collapsed %d request-body oneOf wrapper(s), rewrote %d apiKey scheme(s) to http/bearer\n", fixed, secFixed)
	return nil
}

// fixSecuritySchemes rewrites every apiKey-in-`Authorization`-header security
// scheme into a proper HTTP Bearer scheme (type: http, scheme: bearer). swag v2
// cannot emit `type: http` from annotations, so the generated spec describes the
// JWT as a raw apiKey header — which makes importers (Postman "API Key" auth)
// send the bare token without the "Bearer " prefix the auth middleware requires.
// It mutates doc in place and returns the number of schemes rewritten.
func fixSecuritySchemes(doc map[string]any) int {
	components, ok := doc["components"].(map[string]any)
	if !ok {
		return 0
	}
	schemes, ok := components["securitySchemes"].(map[string]any)
	if !ok {
		return 0
	}

	count := 0
	for name, raw := range schemes {
		scheme, ok := raw.(map[string]any)
		if !ok || !isAuthorizationAPIKey(scheme) {
			continue
		}
		schemes[name] = map[string]any{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		}
		count++
	}
	return count
}

// isAuthorizationAPIKey reports whether scheme is an apiKey carried in the
// `Authorization` request header — the shape swag emits for a JWT bearer token.
func isAuthorizationAPIKey(scheme map[string]any) bool {
	t, _ := scheme["type"].(string)
	in, _ := scheme["in"].(string)
	name, _ := scheme["name"].(string)
	return t == "apiKey" && in == "header" && strings.EqualFold(name, "Authorization")
}

// fixRequestBodies walks every operation's request body and replaces a buggy
// swag-v2 `oneOf` schema with the plain `$ref` it wraps. It mutates doc in place
// and returns the number of request bodies fixed. The walk is generic over all
// paths/methods/media types, so future endpoints are covered automatically.
func fixRequestBodies(doc map[string]any) int {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return 0
	}

	count := 0
	for _, item := range paths {
		methods, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, op := range methods {
			operation, ok := op.(map[string]any)
			if !ok {
				continue
			}
			body, ok := operation["requestBody"].(map[string]any)
			if !ok {
				continue
			}
			content, ok := body["content"].(map[string]any)
			if !ok {
				continue
			}
			for _, media := range content {
				mediaObj, ok := media.(map[string]any)
				if !ok {
					continue
				}
				schema, ok := mediaObj["schema"].(map[string]any)
				if !ok {
					continue
				}
				if ref := refFromOneOf(schema); ref != "" {
					mediaObj["schema"] = map[string]any{"$ref": ref}
					count++
				}
			}
		}
	}
	return count
}

// refFromOneOf returns the `$ref` carried by a swag-v2 `oneOf` request-body
// schema, or "" if the schema is not such a wrapper.
func refFromOneOf(schema map[string]any) string {
	members, ok := schema["oneOf"].([]any)
	if !ok {
		return ""
	}
	for _, m := range members {
		member, ok := m.(map[string]any)
		if !ok {
			continue
		}
		if ref, ok := member["$ref"].(string); ok && ref != "" {
			return ref
		}
	}
	return ""
}
