package datastore

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	"kyouen-server/internal/config"
)

// FirebaseService handles Firebase Authentication operations
type FirebaseService struct {
	auth *auth.Client
}

// NewFirebaseService creates a new Firebase service instance
func NewFirebaseService(cfg *config.Config) (*FirebaseService, error) {
	ctx := context.Background()
	
	var opts []option.ClientOption
	if cfg.FirebaseConfig.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.FirebaseConfig.CredentialsFile))
	}
	
	// Initialize Firebase app
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: cfg.ProjectID,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	
	// Get Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}
	
	return &FirebaseService{
		auth: authClient,
	}, nil
}

// VerifyIDToken verifies Firebase ID token and returns the token claims
func (fs *FirebaseService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := fs.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	
	return token, nil
}

// GetUserByUID retrieves user information from Firebase Auth
func (fs *FirebaseService) GetUserByUID(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := fs.auth.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by UID: %w", err)
	}
	
	return user, nil
}

// DeleteUser deletes a user from Firebase Auth
func (fs *FirebaseService) DeleteUser(ctx context.Context, uid string) error {
	err := fs.auth.DeleteUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to delete user from Firebase Auth: %w", err)
	}
	
	return nil
}