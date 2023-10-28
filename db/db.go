package db

import (
	"context"
	"os"

	"cloud.google.com/go/datastore"
)

var client *datastore.Client

func InitDB() {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	c, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	client = c
}

func DB() *datastore.Client {
	return client
}
