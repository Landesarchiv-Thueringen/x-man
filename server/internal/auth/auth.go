package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthRequired requires a valid bearer token or else stops the request.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		result, err := validateToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Set("userId", result.UserID)
	}
}

func AuthRequiredQueryParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		result, err := validateToken(token)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Set("userId", result.UserID)
	}
}

// AdminRequired requires a valid bearer token with permission "admin" or else stops the request.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		result, err := validateToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else if !result.Permissions.Admin {
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

// Login reads LDAP username and password from HTTP basic auth and responds with
// a bearer token and information about permissions and the user.
func Login(c *gin.Context) {
	user, pass, ok := c.Request.BasicAuth()
	if !ok {
		c.String(http.StatusUnauthorized, "Basic Authentication must be provided")
		return
	}
	authorization := authorizeUser(user, pass)
	switch authorization.Predicate {
	case INVALID:
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	case DENIED:
		c.String(http.StatusForbidden, "Forbidden")
		return
	case GRANTED:
		// continue
	default:
		panic(fmt.Sprintf("unknown authorization predicate: %v", authorization.Predicate))
	}
	token := createToken(*authorization.UserEntry)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  authorization.UserEntry,
	})
}

// extractToken reads a bearer token from the HTTP authorization header.
func extractToken(c *gin.Context) string {
	authorization := c.Request.Header.Get("Authorization")
	if split := strings.Split(authorization, " "); len(split) == 2 && split[0] == "Bearer" {
		return split[1]
	}
	return ""
}
