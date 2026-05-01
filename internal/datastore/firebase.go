package datastore

import (
	"context"
	"encoding/json"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"kyouen-server/internal/config"
)

// FirebaseService handles Firebase Authentication and Messaging operations
type FirebaseService struct {
	auth      *auth.Client
	messaging *messaging.Client
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

	// Get Messaging client
	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Messaging client: %w", err)
	}

	return &FirebaseService{
		auth:      authClient,
		messaging: messagingClient,
	}, nil
}

// NewStageTopic is the FCM topic for new stage notifications.
const NewStageTopic = "stage_added"

// SendNewStageNotification sends a localized FCM push notification to the new-stage topic.
// stageNos is included as a JSON array in the message data so clients can display stage numbers.
func (fs *FirebaseService) SendNewStageNotification(ctx context.Context, stageNos []int64) error {
	stageNosJSON, err := json.Marshal(stageNos)
	if err != nil {
		return fmt.Errorf("failed to marshal stage nos: %w", err)
	}

	message := &messaging.Message{
		Topic: NewStageTopic,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				TitleLocKey: "notification_new_stage_title",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						TitleLocKey: "notification_new_stage_title",
					},
				},
			},
		},
		Data: map[string]string{
			"stage_nos": string(stageNosJSON),
		},
	}

	_, err = fs.messaging.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send FCM notification to topic %s: %w", NewStageTopic, err)
	}
	return nil
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