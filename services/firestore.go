package services

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"kyouen-server/config"
)

type FirestoreService struct {
	client *firestore.Client
	ctx    context.Context
}

func NewFirestoreService(cfg *config.Config) (*FirestoreService, error) {
	ctx := context.Background()
	
	client, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, err
	}

	return &FirestoreService{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *FirestoreService) Close() error {
	return s.client.Close()
}

func (s *FirestoreService) GetClient() *firestore.Client {
	return s.client
}

// Collection helpers
func (s *FirestoreService) Collection(name string) *firestore.CollectionRef {
	return s.client.Collection(name)
}

// Document helpers
func (s *FirestoreService) Doc(collectionName, docID string) *firestore.DocumentRef {
	return s.client.Collection(collectionName).Doc(docID)
}

// Common CRUD operations
func (s *FirestoreService) Create(collectionName string, data interface{}) (*firestore.DocumentRef, error) {
	return s.client.Collection(collectionName).NewDoc(), nil
}

func (s *FirestoreService) Get(collectionName, docID string, dest interface{}) error {
	doc, err := s.client.Collection(collectionName).Doc(docID).Get(s.ctx)
	if err != nil {
		return err
	}
	return doc.DataTo(dest)
}

func (s *FirestoreService) List(collectionName string, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	iter := s.client.Collection(collectionName).Limit(limit).Documents(s.ctx)
	defer iter.Stop()
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		
		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
	}
	
	return results, nil
}

func (s *FirestoreService) Update(collectionName, docID string, updates []firestore.Update) error {
	_, err := s.client.Collection(collectionName).Doc(docID).Update(s.ctx, updates)
	return err
}

func (s *FirestoreService) Delete(collectionName, docID string) error {
	_, err := s.client.Collection(collectionName).Doc(docID).Delete(s.ctx)
	return err
}

// Batch operations
func (s *FirestoreService) Batch() *firestore.WriteBatch {
	return s.client.Batch()
}

func (s *FirestoreService) RunTransaction(fn func(ctx context.Context, tx *firestore.Transaction) error) error {
	return s.client.RunTransaction(s.ctx, fn)
}

// Utility functions
func (s *FirestoreService) AddTimestamps(data map[string]interface{}) map[string]interface{} {
	now := time.Now()
	if _, exists := data["createdAt"]; !exists {
		data["createdAt"] = now
	}
	data["updatedAt"] = now
	return data
}