package db

import (
	"context"

	"cloud.google.com/go/datastore"
)

var client *datastore.Client

func InitDB() {
	ctx := context.Background()
	projectID := "my-android-server"
	c, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	client = c
}

func DB() *datastore.Client {
	return client
}
