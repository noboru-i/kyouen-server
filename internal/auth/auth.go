package auth

import (
	"context"
	"errors"
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
	// GuestUID is the special UID used for guest account (matches existing production data)
	GuestUID = "0"
)

// AuthenticatedUser represents the authenticated user information
type AuthenticatedUser struct {
	UID        string
	Email      string
	Name       string
	Picture    string
	TwitterUID string // Twitter User ID from custom claims
}

// AuthResult represents the result of authentication attempt
type AuthResult struct {
	Success bool
	User    *AuthenticatedUser
	UID     string
	Error   error
}

// authenticateToken performs Firebase token authentication
func authenticateToken(firebaseService *datastore.FirebaseService, authHeader string) *AuthResult {
	if authHeader == "" {
		return &AuthResult{Success: false, Error: errors.New("authorization header is required")}
	}

	// Check Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return &AuthResult{Success: false, Error: errors.New("authorization header must be Bearer token")}
	}

	idToken := parts[1]

	// Verify Firebase ID token
	ctx := context.Background()
	token, err := firebaseService.VerifyIDToken(ctx, idToken)
	if err != nil {
		return &AuthResult{Success: false, Error: errors.New("invalid Firebase ID token")}
	}

	// Get user information from Firebase Auth
	userRecord, err := firebaseService.GetUserByUID(ctx, token.UID)
	if err != nil {
		return &AuthResult{Success: false, Error: errors.New("failed to get user information")}
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

	return &AuthResult{Success: true, User: authUser, UID: token.UID}
}

// setAuthContextFromResult sets authentication information in gin context
func setAuthContextFromResult(c *gin.Context, authResult *AuthResult) {
	if authResult.Success {
		c.Set(AuthUserKey, authResult.User)
		c.Set(AuthUIDKey, authResult.UID)
	} else {
		c.Set(AuthUIDKey, GuestUID)
	}
}

// FirebaseAuth creates a middleware that validates Firebase ID tokens
func FirebaseAuth(firebaseService *datastore.FirebaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authResult := authenticateToken(firebaseService, c.GetHeader("Authorization"))

		if !authResult.Success {
			var statusCode int
			if authResult.Error.Error() == "failed to get user information" {
				statusCode = http.StatusInternalServerError
			} else {
				statusCode = http.StatusUnauthorized
			}

			c.JSON(statusCode, gin.H{"error": authResult.Error.Error()})
			c.Abort()
			return
		}

		setAuthContextFromResult(c, authResult)
		c.Next()
	}
}

// OptionalFirebaseAuth creates a middleware that validates Firebase ID tokens but allows guest access
func OptionalFirebaseAuth(firebaseService *datastore.FirebaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authResult := authenticateToken(firebaseService, c.GetHeader("Authorization"))
		setAuthContextFromResult(c, authResult)

		// Continue to next handler (never abort)
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

// IsGuestUser checks if the current user is a guest user
func IsGuestUser(uid string) bool {
	return uid == GuestUID
}
