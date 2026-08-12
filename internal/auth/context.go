package auth

import "github.com/gin-gonic/gin"

const (
	ContextAuthMethodKey = "authMethod"
	ContextPATIDKey      = "patID"
	ContextPATScopesKey  = "patScopes"

	AuthMethodJWT = "jwt"
	AuthMethodPAT = "pat"
)

func AuthMethod(c *gin.Context) string {
	v, _ := c.Get(ContextAuthMethodKey)
	s, _ := v.(string)
	return s
}

func IsPAT(c *gin.Context) bool {
	return AuthMethod(c) == AuthMethodPAT
}
