package main

// note that this uses data from the seed
// just spamming out int test against a test db because i'm in a rush
// but obviously i'd want unit tests with each layer mocked out in a real codebase
// integration tests like this are still very important but they should be in addition to unit tests, not instead of them
// another note is that this test is brittle because of manually managing the sql row names and the fact that it relies on the seed function in the repo which could change, ideally we'd want to insert known test data in the test itself and then query against that
// so it's testing behaviour more than hyper specific data which could change
// considering the proto gen stuff glues right against the service call, i don't think it's worth testing that part of the integration -
// that would belong in an e2e or smoke test

import (
	"context"
	"database/sql"
	"testing"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"git.neds.sh/matty/entain/racing/service"
	_ "github.com/mattn/go-sqlite3"
)

//
// list races
//

func TestListRaces_FilterFlowsThroughLayers(t *testing.T) {
	racingDb := setupTestDB(t)

	repo := db.NewRacesRepo(racingDb)
	err := repo.Init()
	if err != nil {
		t.Fatal(err)
	}

	svc := service.NewRacingService(repo)

	visible := true

	req := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			Visible: &visible,
		},
	}

	resp, err := svc.ListRaces(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, race := range resp.Races {
		if !race.Visible {
			t.Fatalf("expected only visible races")
		}
	}
}

// test name desc, if that works, the rest probably do too
// again, i'd be more thorough with unit tests in a real codebase, this is just covering the fundamental path sparingly
func TestListRaces_SortByNameDescFlowsThroughLayers(t *testing.T) {
	racingDb := setupTestDB(t)

	repo := db.NewRacesRepo(racingDb)
	if err := repo.Init(); err != nil {
		t.Fatal(err)
	}

	svc := service.NewRacingService(repo)

	req := &racing.ListRacesRequest{
		OrderBy: "name DESC",
	}

	resp, err := svc.ListRaces(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Races) < 2 {
		t.Fatalf("expected seeded data to have multiple races")
	}

	// verify ordering is descending by name
	for i := 1; i < len(resp.Races); i++ {
		prev := resp.Races[i-1].Name
		curr := resp.Races[i].Name

		if prev < curr {
			t.Fatalf("expected DESC order, got %s before %s", prev, curr)
		}
	}
}

//
// get race
//

func TestGetRace_FlowsThroughLayers(t *testing.T) {
	racingDb := setupTestDB(t)

	repo := db.NewRacesRepo(racingDb)
	if err := repo.Init(); err != nil {
		t.Fatal(err)
	}

	// insert known test data (avoid seed reliance)
	_, err := racingDb.Exec(`
		INSERT INTO races(id, meeting_id, name, number, visible, advertised_start_time)
		VALUES (123, 1, 'Test Race', 1, 1, datetime('now', '+1 day'))
	`)
	if err != nil {
		t.Fatal(err)
	}

	svc := service.NewRacingService(repo)

	req := &racing.GetRaceRequest{
		Id: 123,
	}

	resp, err := svc.GetRace(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Race == nil {
		t.Fatal("expected race, got nil")
	}

	if resp.Race.Id != 123 {
		t.Fatalf("expected id 123, got %d", resp.Race.Id)
	}

	if resp.Race.Name != "Test Race" {
		t.Fatalf("expected name 'Test Race', got %s", resp.Race.Name)
	}

	if resp.Race.Visible != true {
		t.Fatalf("expected visible true")
	}
}

//
// helpers
//

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
