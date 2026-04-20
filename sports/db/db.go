package db

import (
	"fmt"
	"time"

	"syreclabs.com/go/faker"
)

func (s *eventsRepo) seed() error {
	statement, err := s.db.Prepare(`
		CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY,
			name TEXT,
			sport_type INTEGER,
			competitors TEXT,
			advertised_start_time DATETIME
		)
	`)
	if err != nil {
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		return err
	}

	sports := []int{1, 2, 3, 4} // matches enum values

	for i := 1; i <= 100; i++ {
		statement, err := s.db.Prepare(`
			INSERT OR IGNORE INTO events(
				id,
				name,
				sport_type,
				competitors,
				advertised_start_time
			) VALUES (?,?,?,?,?)
		`)
		if err != nil {
			return err
		}

		teamA := faker.Team().Name()
		teamB := faker.Team().Name()

		name := fmt.Sprintf("%s vs %s", teamA, teamB)

		competitors := fmt.Sprintf("%s,%s", teamA, teamB)

		_, err = statement.Exec(
			i,
			name,
			faker.Number().Between(0, len(sports)-1),
			competitors,
			faker.Time().
				Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).
				Format(time.RFC3339),
		)

		if err != nil {
			return err
		}
	}

	return nil
}
