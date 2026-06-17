package token

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateAndParse_RoundTrip — токен, выпущенный менеджером, парсится им же
// и отдаёт исходные claims; expiresIn равен TTL в секундах.
func TestGenerateAndParse_RoundTrip(t *testing.T) {
	mgr := NewManager("test-secret", 15*time.Minute)

	signed, expiresIn, err := mgr.GenerateAccess(42, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, int64((15 * time.Minute).Seconds()), expiresIn)

	claims, err := mgr.ParseAccess(signed)
	require.NoError(t, err)
	assert.Equal(t, "42", claims.Subject)
	assert.Equal(t, "user@example.com", claims.Email)
}

// TestParseAccess_WrongSecret — токен, подписанный другим секретом, отвергается.
func TestParseAccess_WrongSecret(t *testing.T) {
	issuer := NewManager("secret-a", time.Minute)
	verifier := NewManager("secret-b", time.Minute)

	signed, _, err := issuer.GenerateAccess(1, "u@e.com")
	require.NoError(t, err)

	_, err = verifier.ParseAccess(signed)
	assert.Error(t, err)
}

// TestParseAccess_Expired — истёкший токен не проходит проверку срока действия.
func TestParseAccess_Expired(t *testing.T) {
	mgr := NewManager("test-secret", -time.Minute) // exp в прошлом

	signed, _, err := mgr.GenerateAccess(1, "u@e.com")
	require.NoError(t, err)

	_, err = mgr.ParseAccess(signed)
	assert.Error(t, err)
}

// TestParseAccess_TamperedSignature — подмена байта в подписи ломает валидацию.
func TestParseAccess_TamperedSignature(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)

	signed, _, err := mgr.GenerateAccess(1, "u@e.com")
	require.NoError(t, err)

	parts := strings.Split(signed, ".")
	require.Len(t, parts, 3)
	// Меняем последний символ подписи на заведомо другой.
	sig := parts[2]
	last := sig[len(sig)-1]
	repl := byte('A')
	if last == repl {
		repl = 'B'
	}
	parts[2] = sig[:len(sig)-1] + string(repl)
	tampered := strings.Join(parts, ".")

	_, err = mgr.ParseAccess(tampered)
	assert.Error(t, err)
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
	require.NoError(t, err)

	_, err = mgr.ParseAccess(noneToken)
	assert.Error(t, err)
}

// TestParseAccess_Garbage — нераспознаваемая строка даёт ошибку и nil claims.
func TestParseAccess_Garbage(t *testing.T) {
	mgr := NewManager("test-secret", time.Minute)

	claims, err := mgr.ParseAccess("not.a.jwt")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
