package stage

import (
	"testing"

	"cloud.google.com/go/datastore"
)

func makeKey(kind string, id int64) *datastore.Key {
	return &datastore.Key{Kind: kind, ID: id}
}

func TestUniqueKeys_Deduplication(t *testing.T) {
	k1 := makeKey("User", 1)
	k2 := makeKey("User", 2)
	keys := []*datastore.Key{k1, k2, k1, k2, k1}

	unique, idx := uniqueKeys(keys)

	if len(unique) != 2 {
		t.Errorf("Expected 2 unique keys, got %d", len(unique))
	}
	if len(idx) != 2 {
		t.Errorf("Expected idx length 2, got %d", len(idx))
	}
	if idx[k1.String()] != 0 {
		t.Errorf("Expected k1 at index 0, got %d", idx[k1.String()])
	}
	if idx[k2.String()] != 1 {
		t.Errorf("Expected k2 at index 1, got %d", idx[k2.String()])
	}
}

func TestUniqueKeys_Empty(t *testing.T) {
	unique, idx := uniqueKeys([]*datastore.Key{})

	if len(unique) != 0 {
		t.Errorf("Expected 0 unique keys, got %d", len(unique))
	}
	if len(idx) != 0 {
		t.Errorf("Expected empty idx, got %d", len(idx))
	}
}

func TestUniqueKeys_OrderPreserved(t *testing.T) {
	k1 := makeKey("Stage", 10)
	k2 := makeKey("Stage", 20)
	k3 := makeKey("Stage", 30)
	keys := []*datastore.Key{k3, k1, k2, k3, k1}

	unique, _ := uniqueKeys(keys)

	if len(unique) != 3 {
		t.Errorf("Expected 3 unique keys, got %d", len(unique))
	}
	if unique[0].ID != 30 || unique[1].ID != 10 || unique[2].ID != 20 {
		t.Errorf("Expected order [k3, k1, k2], got [%d, %d, %d]", unique[0].ID, unique[1].ID, unique[2].ID)
	}
}
