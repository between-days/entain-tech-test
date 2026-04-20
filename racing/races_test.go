package main

// just spamming out int test against a test db because i'm in a rush
// but obviously i'd want unit tests with each layer mocked out in a real codebase
// integration tests like this are still very important but they should be in addition to unit tests, not instead of them
// another note is that this test is brittle because of manually managing the sql row names and the fact that it relies on the seed function in the repo which could change, ideally we'd want to insert known test data in the test itself and then query against that
// so it's testing behaviour more than hyper specific data which could change

import (
	"database/sql"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	_ "github.com/mattn/go-sqlite3"
)

func TestListRaces_NoFilterReturnsAll(t *testing.T) {
	racingDb := setupTestDB(t)
	repo := db.NewRacesRepo(racingDb)

	insertRace(t, racingDb, 1, true)
	insertRace(t, racingDb, 2, false)

	races, err := repo.List(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(races) != 2 {
		t.Fatalf("expected 2 races, got %d", len(races))
	}
}

func TestListRaces_VisibleTrueFilters(t *testing.T) {
	racingDb := setupTestDB(t)
	repo := db.NewRacesRepo(racingDb)

	insertRace(t, racingDb, 1, true)
	insertRace(t, racingDb, 2, false)

	visible := true
	filter := &racing.ListRacesRequestFilter{
		Visible: &visible,
	}

	races, err := repo.List(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(races))
	}

	for _, r := range races {
		if !r.Visible {
			t.Fatalf("expected only visible races")
		}
	}
}

func TestListRaces_VisibleFalseFilters(t *testing.T) {
	racingDb := setupTestDB(t)
	repo := db.NewRacesRepo(racingDb)

	insertRace(t, racingDb, 1, true)
	insertRace(t, racingDb, 2, false)

	visible := false
	filter := &racing.ListRacesRequestFilter{
		Visible: &visible,
	}

	races, err := repo.List(filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(races))
	}

	for _, r := range races {
		if r.Visible {
			t.Fatalf("expected only non-visible races")
		}
	}
}

// helpers

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// create schema (copy from your seed)
	_, err = db.Exec(`
		CREATE TABLE races (
			id INTEGER PRIMARY KEY,
			meeting_id INTEGER,
			name TEXT,
			number INTEGER,
			visible INTEGER,
			advertised_start_time DATETIME
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func insertRace(t *testing.T, db *sql.DB, id int, visible bool) {
	_, err := db.Exec(`
		INSERT INTO races(id, meeting_id, name, number, visible, advertised_start_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		id,
		1,
		"test race",
		1,
		boolToInt(visible),
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
