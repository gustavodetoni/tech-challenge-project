package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	auth2 "github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
	"github.com/stretchr/testify/require"
)

func TestManager_NewToken_RequiresSecretAndSubject(t *testing.T) {
	m := auth2.NewManager("", "issuer", "", time.Minute)
	_, _, err := m.NewToken("sub", "ADMIN")
	require.Error(t, err)

	m = auth2.NewManager("secret", "issuer", "", time.Minute)
	_, _, err = m.NewToken("", "ADMIN")
	require.Error(t, err)
}

func TestManager_ParseAndValidate_Errors(t *testing.T) {
	m := auth2.NewManager("", "issuer", "", time.Minute)
	_, err := m.ParseAndValidate("token")
	require.Error(t, err)

	m = auth2.NewManager("secret", "issuer", "", time.Minute)
	_, err = m.ParseAndValidate("")
	require.Error(t, err)

	// Wrong signing method (HS512).
	claims := auth2.Claims{Role: "ADMIN", RegisteredClaims: jwt.RegisteredClaims{Subject: "sub"}}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	signed, signErr := tok.SignedString([]byte("secret"))
	require.NoError(t, signErr)

	_, err = m.ParseAndValidate(signed)
	require.Error(t, err)
}

func TestManager_NewToken_SetsAudienceWhenProvided(t *testing.T) {
	m := auth2.NewManager("secret", "issuer", "aud", time.Minute)
	token, _, err := m.NewToken("sub", "ADMIN")
	require.NoError(t, err)

	claims, err := m.ParseAndValidate(token)
	require.NoError(t, err)
	require.Len(t, claims.Audience, 1)
	require.Equal(t, "aud", claims.Audience[0])
}
