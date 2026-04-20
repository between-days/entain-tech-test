package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// EventsRepo provides repository access to events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return a list of events.
	List() ([]*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

func (r *eventsRepo) List() ([]*sports.Event, error) {
	rows, err := r.db.Query(getEventQueries()[eventsList])
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows)
}

func toSportType(v int32) sports.SportType {
	st := sports.SportType(v)
	if _, ok := sports.SportType_name[int32(st)]; !ok {
		return sports.SportType_UNSPECIFIED
	}
	return st
}

func (r *eventsRepo) scanEvents(rows *sql.Rows) ([]*sports.Event, error) {
	var events []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time
		var sportTypeInt int32
		var competitorsStr string

		if err := rows.Scan(&event.Id, &event.Name, &sportTypeInt, &competitorsStr, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		event.SportType = toSportType(sportTypeInt)

		if competitorsStr != "" {
			event.Competitors = strings.Split(competitorsStr, ",")
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts

		events = append(events, &event)
	}

	return events, nil
}
