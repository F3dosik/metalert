package main

import (
	"testing"
)

func TestUserReset(t *testing.T) {
	zip := "12345"
	u := &User{
		ID:       42,
		Name:     "Alice",
		Email:    "alice@example.com",
		Active:   true,
		Score:    9.8,
		Tags:     []string{"admin", "user"},
		Metadata: map[string]string{"key": "val"},
		Address: &Address{
			Street:  "Main St",
			City:    "Springfield",
			Country: "US",
			Zip:     &zip,
			Coords:  []float64{1.1, 2.2},
		},
		Profile: Profile{Bio: "bio", Avatar: "img.png"},
	}

	u.Reset()

	if u.ID != 0 {
		t.Errorf("ID: want 0, got %d", u.ID)
	}
	if u.Name != "" {
		t.Errorf("Name: want empty, got %q", u.Name)
	}
	if u.Active {
		t.Error("Active: want false")
	}
	if u.Score != 0 {
		t.Errorf("Score: want 0, got %f", u.Score)
	}
	if len(u.Tags) != 0 {
		t.Errorf("Tags len: want 0, got %d", len(u.Tags))
	}
	if cap(u.Tags) == 0 {
		t.Error("Tags cap: should be preserved after [:0]")
	}
	if len(u.Metadata) != 0 {
		t.Errorf("Metadata len: want 0, got %d", len(u.Metadata))
	}
	// Address has Reset() — должен быть вызван
	if u.Address == nil {
		t.Fatal("Address pointer must not be nil after Reset")
	}
	if u.Address.Street != "" {
		t.Errorf("Address.Street: want empty, got %q", u.Address.Street)
	}
	if *u.Address.Zip != "" {
		t.Errorf("Address.Zip value: want empty, got %q", *u.Address.Zip)
	}
}

func TestNilSafe(t *testing.T) {
	var u *User
	// Не должно паниковать
	u.Reset()
	_ = u
}
