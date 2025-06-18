package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"kyouen-server/internal/datastore"
	"kyouen-server/pkg/models"
)

type SeedStage struct {
	StageNo    string `json:"stageNo"`
	Size       string `json:"size"`
	RegistDate string `json:"registDate"`
	Stage      string `json:"stage"`
	Creator    string `json:"creator"`
}

var seedData = []SeedStage{
	{"1", "6", "None", "000000010000001100001100000000001000", "noboru"},
	{"2", "6", "None", "000000000000000100010010001100000000", "noboru"},
	{"3", "6", "None", "000000001000010000000100010010001000", "noboru"},
	{"4", "6", "None", "001000001000000010010000010100000000", "noboru"},
	{"5", "6", "None", "000000001011010000000010001000000010", "noboru"},
	{"6", "6", "None", "000100000000101011010000000000000000", "noboru"},
	{"7", "6", "None", "000000001010000000010010000000001010", "noboru"},
	{"8", "6", "None", "001000000001010000010010000001000000", "noboru"},
	{"9", "6", "None", "000000001000010000000010000100001000", "noboru"},
	{"10", "6", "None", "000100000010010000000100000010010000", "noboru"},
}

func main() {
	ctx := context.Background()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "my-android-server"
		log.Printf("Using default project ID: %s", projectID)
	}

	datastoreService, err := datastore.NewDatastoreService(projectID)
	if err != nil {
		log.Fatal("Failed to create datastore service:", err)
	}
	defer datastoreService.Close()

	fmt.Println("Starting seed data initialization...")

	for _, seed := range seedData {
		if err := createStage(ctx, datastoreService, seed); err != nil {
			log.Printf("Failed to create stage %s: %v", seed.StageNo, err)
			continue
		}
		fmt.Printf("Created stage %s successfully\n", seed.StageNo)
	}

	fmt.Println("Seed data initialization completed")
}

func createStage(_ context.Context, datastoreService *datastore.DatastoreService, seed SeedStage) error {
	size, err := strconv.ParseInt(seed.Size, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid size: %s", seed.Size)
	}

	kyouenStage := models.NewKyouenStage(int(size), seed.Stage)
	if kyouenStage.StoneCount() <= 4 {
		return fmt.Errorf("invalid stage: insufficient stones")
	}
	if kyouenStage.HasKyouen() == nil {
		return fmt.Errorf("invalid stage: does not contain valid kyouen")
	}

	exists, err := datastoreService.CheckStageExists(seed.Stage)
	if err != nil {
		return fmt.Errorf("failed to check stage existence: %w", err)
	}
	if exists {
		fmt.Printf("Stage %s already exists, skipping\n", seed.StageNo)
		return nil
	}

	stage := datastore.KyouenPuzzle{
		Size:       size,
		Stage:      seed.Stage,
		Creator:    seed.Creator,
		RegistDate: time.Now(),
	}

	_, err = datastoreService.CreateStage(stage)
	if err != nil {
		return fmt.Errorf("failed to create stage: %w", err)
	}

	return nil
}
