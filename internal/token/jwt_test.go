package token

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateAndParse_RoundTrip — токен, выпущенный менеджером, парсится им же
// и отдаёт исходные claims; expiresIn равен TTL в секундах.
func TestGenerateAndParse_RoundTrip(t *testing.T) {
	mgr := NewManager("test-secret", 15*time.Minute)

	signed, expiresIn, err := mgr.GenerateAccess(42, "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expiresIn != int64((15 * time.Minute).Seconds()) {
		t.Errorf("expected expiresIn=%d, got %d", int64((15 * time.Minute).Seconds()), expiresIn)
	}

	claims, err := mgr.ParseAccess(signed)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if claims.Subject != "42" {
		t.Errorf("expected subject %q, got %q", "42", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("expected email %q, got %q", "user@example.com", claims.Email)
	}
}

// TestParseAccess_WrongSecret — токен, подписанный другим секретом, отвергается.
func TestParseAccess_WrongSecret(t *testing.T) {
	issuer := NewManager("secret-a", time.Minute)
	verifier := NewManager("secret-b", time.Minute)

	signed, _, err := issuer.GenerateAccess(1, "u@e.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := verifier.ParseAccess(signed); err == nil {
		t.Error("expected error parsing token signed with a different secret, got nil")
	}
}

// TestParseAccess_Expired — истёкший токен не проходит проверку срока действия.
func TestParseAccess_Expired(t *testing.T) {
	mgr := NewManager("test-secret", -time.Minute) // exp в прошлом

	signed, _, err := mgr.GenerateAccess(1, "u@e.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := mgr.ParseAccess(signed); err == nil {
		t.Error("expected error parsing expired token, got nil")
	}
}

// TestParseAccess_TamperedSignature — подмена байта в подписи ломает валидацию.
func TestParseAccess_TamperedSignature(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)

	signed, _, err := mgr.GenerateAccess(1, "u@e.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(signed, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT segments, got %d", len(parts))
	}
	// Меняем последний символ подписи на заведомо другой.
	sig := parts[2]
	last := sig[len(sig)-1]
	repl := byte('A')
	if last == repl {
		repl = 'B'
	}
	parts[2] = sig[:len(sig)-1] + string(repl)
	tampered := strings.Join(parts, ".")

	if _, err := mgr.ParseAccess(tampered); err == nil {
		t.Error("expected error parsing token with tampered signature, got nil")
	}
}

// TestParseAccess_AlgConfusion — токен с алгоритмом "none" отвергается:
// keyfunc требует именно *jwt.SigningMethodHMAC (см. jwt.go).
func TestParseAccess_AlgConfusion(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	noneToken, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to build none-signed token: %v", err)
	}

	if _, err := mgr.ParseAccess(noneToken); err == nil {
		t.Error("expected alg-confusion (none) token to be rejected, got nil")
	}
}

// TestParseAccess_Garbage — нераспознаваемая строка даёт ошибку и nil claims.
func TestParseAccess_Garbage(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)

	claims, err := mgr.ParseAccess("not.a.jwt")
	if err == nil {
		t.Error("expected error parsing garbage token, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims on error, got %+v", claims)
	}
}
