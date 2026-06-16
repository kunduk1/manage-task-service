// Command swagfix post-processes the OpenAPI spec emitted by swag v2 to work
// around https://github.com/swaggo/swag/issues/2086: swag wraps every request
// body in a malformed `oneOf` whose first member is an empty `{type: object}`.
// Clients like Postman generate their example body from that first member, so
// the body comes out empty (and, with no body, they drop the Content-Type
// header). We collapse the oneOf back to the real `$ref` it wraps so the
// per-field `example` values are used.
//
// It rewrites <dir>/swagger.json and <dir>/swagger.yaml (dir defaults to
// "docs"). docs/docs.go is intentionally left untouched: the service serves the
// embedded swagger.json directly and bypasses the swag registry (see
// docs/embed.go), so the template there is unused at runtime.
//
// The fix is idempotent — once the oneOf is gone, re-running is a no-op.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	fmt.Printf("swagfix: collapsed %d request-body oneOf wrapper(s)\n", fixed)
	return nil
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
