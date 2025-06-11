package testutils

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/datastore"
)

func SetupDatastoreTest() *datastore.Client {
	// Set emulator host for testing
	os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")
	
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "test-project")
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func CleanupDatastore(client *datastore.Client) {
	ctx := context.Background()
	entities := []string{"KyouenPuzzle", "User", "StageUser", "KyouenPuzzleSummary"}
	
	for _, entityKind := range entities {
		query := datastore.NewQuery(entityKind).KeysOnly()
		keys, err := client.GetAll(ctx, query, nil)
		if err != nil {
			log.Printf("Error getting keys for %s: %v", entityKind, err)
			continue
		}
		
		if len(keys) > 0 {
			err = client.DeleteMulti(ctx, keys)
			if err != nil {
				log.Printf("Error cleaning up %s: %v", entityKind, err)
			}
		}
	}
}