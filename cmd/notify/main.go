package main

import (
	"context"
	"log"

	"kyouen-server/internal/config"
	"kyouen-server/internal/datastore"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	datastoreService, err := datastore.NewDatastoreService(cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize Datastore service: %v", err)
	}
	defer datastoreService.Close()

	firebaseService, err := datastore.NewFirebaseService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase service: %v", err)
	}

	stageKeys, err := datastoreService.GetAndDeleteRegistModels()
	if err != nil {
		log.Fatalf("Failed to get RegistModels: %v", err)
	}

	if len(stageKeys) == 0 {
		log.Printf("No new stages registered, skipping notification")
		return
	}

	stages, err := datastoreService.GetStagesByKeys(stageKeys)
	if err != nil {
		log.Fatalf("Failed to get stages: %v", err)
	}

	stageNos := make([]int64, 0, len(stages))
	for _, s := range stages {
		if s.StageNo > 0 {
			stageNos = append(stageNos, s.StageNo)
		}
	}

	log.Printf("Found %d new stage(s) %v, sending push notification to topic: %s", len(stageNos), stageNos, datastore.NewStageTopic)

	if err := firebaseService.SendNewStageNotification(ctx, stageNos); err != nil {
		log.Fatalf("Failed to send notification: %v", err)
	}

	log.Printf("Push notification sent successfully")
}
