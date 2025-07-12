package auth

import (
	"context"
	"net/http"
	"strings"

	"kyouen-server/internal/datastore"

	"github.com/gin-gonic/gin"
)

const (
	// AuthUserKey is the key used to store authenticated user info in gin context
	AuthUserKey = "auth_user"
	// AuthUIDKey is the key used to store authenticated user's Firebase UID in gin context
	AuthUIDKey = "auth_uid"
)

// AuthenticatedUser represents the authenticated user information
type AuthenticatedUser struct {
	UID        string
	Email      string
	Name       string
	Picture    string
	TwitterUID string // Twitter User ID from custom claims
}

// FirebaseAuth creates a middleware that validates Firebase ID tokens
func FirebaseAuth(firebaseService *datastore.FirebaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must be Bearer token",
			})
			c.Abort()
			return
		}

		idToken := parts[1]

		// Verify Firebase ID token
		ctx := context.Background()
		token, err := firebaseService.VerifyIDToken(ctx, idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Firebase ID token",
			})
			c.Abort()
			return
		}

		// Get user information from Firebase Auth
		userRecord, err := firebaseService.GetUserByUID(ctx, token.UID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get user information",
			})
			c.Abort()
			return
		}

		// Extract Twitter UID from custom claims (if available)
		twitterUID := ""
		if claims, ok := token.Claims["firebase"].(map[string]interface{}); ok {
			if identities, ok := claims["identities"].(map[string]interface{}); ok {
				if twitterIds, ok := identities["twitter.com"].([]interface{}); ok && len(twitterIds) > 0 {
					if twitterID, ok := twitterIds[0].(string); ok {
						twitterUID = twitterID
					}
				}
			}
		}

		// Create authenticated user object
		authUser := &AuthenticatedUser{
			UID:        token.UID,
			Email:      userRecord.Email,
			Name:       userRecord.DisplayName,
			Picture:    userRecord.PhotoURL,
			TwitterUID: twitterUID,
		}

		// Store user information in context
		c.Set(AuthUserKey, authUser)
		c.Set(AuthUIDKey, token.UID)

		// Continue to next handler
		c.Next()
	}
}

// GetAuthenticatedUser retrieves the authenticated user from gin context
func GetAuthenticatedUser(c *gin.Context) (*AuthenticatedUser, bool) {
	if user, exists := c.Get(AuthUserKey); exists {
		if authUser, ok := user.(*AuthenticatedUser); ok {
			return authUser, true
		}
	}
	return nil, false
}

// GetAuthenticatedUID retrieves the authenticated user's UID from gin context
func GetAuthenticatedUID(c *gin.Context) (string, bool) {
	if uid, exists := c.Get(AuthUIDKey); exists {
		if uidStr, ok := uid.(string); ok {
			return uidStr, true
		}
	}
	return "", false
}
