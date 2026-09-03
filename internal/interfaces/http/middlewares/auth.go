package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
)

const (
	ContextClaimsKey               = "auth.claims"
	ContextSubjectKey              = "auth.subject"
	ContextClientClaimsKey         = "auth.client_claims"
	ContextClientIDKey             = "auth.client_id"
	ContextClientDocumentNumberKey = "auth.client_document_number"
)

type AuthMiddleware struct {
	jwt       *auth.Manager
	clientJWT *auth.Manager
}

func NewAuthMiddleware(jwt *auth.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, clientJWT: jwt}
}

func NewAuthMiddlewareWithClientJWT(jwt *auth.Manager, clientJWT *auth.Manager) *AuthMiddleware {
	if clientJWT == nil {
		clientJWT = jwt
	}
	return &AuthMiddleware{jwt: jwt, clientJWT: clientJWT}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerTokenFromHeader(c.GetHeader("Authorization"))
		claims, err := m.jwt.ParseAndValidate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(ContextClaimsKey, claims)
		c.Set(ContextSubjectKey, claims.Subject)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		if r == "" {
			continue
		}
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		tokenString := bearerTokenFromHeader(c.GetHeader("Authorization"))
		claims, err := m.jwt.ParseAndValidate(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if len(roleSet) > 0 {
			if _, ok := roleSet[claims.Role]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}

		c.Set(ContextClaimsKey, claims)
		c.Set(ContextSubjectKey, claims.Subject)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireClientAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerTokenFromHeader(c.GetHeader("Authorization"))
		claims, err := m.clientJWT.ParseAndValidate(tokenString)
		if err != nil || !isClientClaims(claims) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(ContextClientClaimsKey, claims)
		c.Set(ContextClientIDKey, claims.Subject)
		c.Set(ContextClientDocumentNumberKey, claims.DocumentNumber)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(ContextClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*auth.Claims)
	return claims, ok
}

func GetClientClaims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(ContextClientClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*auth.Claims)
	return claims, ok
}

func GetClientDocumentNumber(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextClientDocumentNumberKey)
	if !ok {
		return "", false
	}
	doc, ok := v.(string)
	if !ok || strings.TrimSpace(doc) == "" {
		return "", false
	}
	return doc, true
}

func isClientClaims(claims *auth.Claims) bool {
	return claims != nil &&
		claims.TokenType == "CLIENT" &&
		claims.Subject != "" &&
		strings.TrimSpace(claims.DocumentNumber) != ""
}

func bearerTokenFromHeader(header string) string {
	if header == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		trimmed := strings.TrimSpace(header)
		if strings.Count(trimmed, ".") == 2 {
			return trimmed
		}
		return ""
	}
	return strings.TrimSpace(header[len("Bearer "):])
}
