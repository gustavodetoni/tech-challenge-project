package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/soat-architecture/tech-challenge-project/internal/infra/auth"
)

const (
	ContextClaimsKey  = "auth.claims"
	ContextSubjectKey = "auth.subject"
)

type AuthMiddleware struct {
	jwt *auth.Manager
}

func NewAuthMiddleware(jwt *auth.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
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

func GetClaims(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(ContextClaimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*auth.Claims)
	return claims, ok
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
