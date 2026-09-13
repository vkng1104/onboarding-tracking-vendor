package vendor

import (
	"errors"
	"testing"
)

func TestParseVendorID(t *testing.T) {
	validID := "20000000-0000-0000-0000-000000000001"
	id, err := parseVendorID(validID)
	if err != nil {
		t.Fatalf("parse valid vendor id: %v", err)
	}
	if !id.Valid {
		t.Fatal("parsed vendor id is not valid")
	}

	if _, err := parseVendorID("not-a-vendor"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("parse invalid vendor id error = %v, want %v", err, ErrNotFound)
	}
}

func TestParseCoordinatorID(t *testing.T) {
	validID := "10000000-0000-0000-0000-000000000001"
	id, err := parseCoordinatorID(validID)
	if err != nil {
		t.Fatalf("parse valid coordinator id: %v", err)
	}
	if !id.Valid {
		t.Fatal("parsed coordinator id is not valid")
	}

	if _, err := parseCoordinatorID("not-a-coordinator"); err == nil {
		t.Fatal("parse invalid coordinator id error = nil, want error")
	}
}
