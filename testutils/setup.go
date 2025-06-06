package testutils

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func SetupFirestoreTest() *firestore.Client {
	// Set emulator host for testing
	os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
	
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "test-project")
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func CleanupFirestore(client *firestore.Client) {
	ctx := context.Background()
	collections := []string{"stages", "users", "stage_users", "summaries"}
	
	for _, collection := range collections {
		iter := client.Collection(collection).Documents(ctx)
		batch := client.Batch()
		
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Error getting document: %v", err)
				continue
			}
			batch.Delete(doc.Ref)
		}
		
		if _, err := batch.Commit(ctx); err != nil {
			log.Printf("Error cleaning up collection %s: %v", collection, err)
		}
	}
}