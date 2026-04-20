package main

import (
	"context"
	"database/sql"
	"testing"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"git.neds.sh/matty/entain/sports/service"
	_ "github.com/mattn/go-sqlite3"
)

func TestListEvents_FlowsThroughLayers(t *testing.T) {
	sportsDb := setupSportsTestDB(t)

	// controlled test data (no seed dependency)
	_, err := sportsDb.Exec(`
		INSERT INTO events (
			id, name, sport_type, competitors, advertised_start_time
		) VALUES (
			1,
			'Test Event',
			2,
			'Player A,Player B',
			datetime('now', '+1 day')
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	repo := db.NewEventsRepo(sportsDb)
	svc := service.NewSportsService(repo)

	req := &sports.ListEventsRequest{}

	resp, err := svc.ListEvents(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(resp.Events))
	}

	e := resp.Events[0]

	if e.Id != 1 {
		t.Fatalf("expected id 1, got %d", e.Id)
	}

	if e.Name != "Test Event" {
		t.Fatalf("expected name 'Test Event', got %s", e.Name)
	}

	if e.SportType != sports.SportType_TENNIS {
		t.Fatalf("expected TENNIS, got %v", e.SportType)
	}

	if len(e.Competitors) != 2 {
		t.Fatalf("expected 2 competitors, got %d", len(e.Competitors))
	}
}

//
// helpers
//

func setupSportsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE events (
			id INTEGER PRIMARY KEY,
			name TEXT,
			sport_type INTEGER,
			competitors TEXT,
			advertised_start_time DATETIME
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
