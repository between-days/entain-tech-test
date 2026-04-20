package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter, orderBy string) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter, orderBy string) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)
	query, _ = r.applySort(query, orderBy)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		args = append(args, *filter.Visible)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

const (
	fieldAdvertisedStartTime = "advertised_start_time"
	fieldName                = "name"
	fieldNumber              = "number"
	fieldMeetingID           = "meeting_id"
)

// need to break the string like "advertised_start_time asc" into the field name and the order (asc/desc)
// for now we only allow 1d sorting, so we can just split on the first space and take the first part as the field name and the second part as the order (if it exists)
// basically default to sorting by advertised_start_time asc, or x asc where x is the field name without the direction
// not bothering with validation, default fall through should be fine for demo
// future would allow multiple sort fields, but for now we just want to demonstrate the concept of sorting and not get bogged down in the details of parsing a complex sort string
func (r *racesRepo) applySort(query string, orderBy string) (string, []interface{}) {
	// don't want an sql injection vulnerabilty
	whitelist := map[string]bool{
		fieldAdvertisedStartTime: true,
		fieldName:                true,
		fieldNumber:              true,
		fieldMeetingID:           true,
	}

	if orderBy == "" {
		query += " ORDER BY advertised_start_time ASC"
		return query, nil
	}

	parts := strings.SplitN(orderBy, " ", 2)
	fieldName := parts[0]

	if !whitelist[fieldName] {
		query += " ORDER BY advertised_start_time ASC"
		return query, nil
	}

	order := "ASC"
	if len(parts) > 1 {
		order = strings.ToUpper(parts[1])
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
	}

	query += " ORDER BY " + fieldName + " " + order

	return query, nil
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
