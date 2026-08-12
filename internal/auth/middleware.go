package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/auth/pat"
)

// PATAuthenticator validates personal access tokens.
type PATAuthenticator interface {
	Authenticate(rawToken string) (*pat.TokenRecord, error)
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := parts[1]
		if strings.HasPrefix(token, pat.TokenPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextAuthMethodKey, AuthMethodJWT)
		c.Next()
	}
}

// AuthAPI accepts JWT session tokens or scoped personal access tokens.
func AuthAPI(authenticator PATAuthenticator, requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := parts[1]
		if strings.HasPrefix(token, pat.TokenPrefix) {
			if authenticator == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				return
			}
			record, err := authenticator.Authenticate(token)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
				return
			}
			if !pat.HasScope(record.Scopes, requiredScope) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient scope"})
				return
			}
			c.Set(ContextUserIDKey, record.UserID)
			c.Set(ContextAuthMethodKey, AuthMethodPAT)
			c.Set(ContextPATIDKey, record.ID)
			c.Set(ContextPATScopesKey, record.Scopes)
			c.Next()
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextAuthMethodKey, AuthMethodJWT)
		c.Next()
	}
}
