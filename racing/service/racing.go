package service

import (
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
)

type Racing interface {
	// GetRace will return a single race
	GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.GetRaceResponse, error)
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
}

// racingService implements the Racing interface.
type racingService struct {
	racesRepo db.RacesRepo
}

// NewRacingService instantiates and returns a new racingService.
func NewRacingService(racesRepo db.RacesRepo) Racing {
	return &racingService{racesRepo}
}

func (s *racingService) GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.GetRaceResponse, error) {
	race, err := s.racesRepo.Get(in.Id)
	if err != nil {
		return nil, err
	}

	enrichRace(race, time.Now())

	return &racing.GetRaceResponse{Race: race}, nil
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	races, err := s.racesRepo.List(in.Filter, in.OrderBy)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for _, r := range races {
		enrichRace(r, now)
	}

	return &racing.ListRacesResponse{Races: races}, nil
}

// derive fields
// for now just derive the status of the race
func enrichRace(r *racing.Race, now time.Time) *racing.Race {
	if now.Before(r.AdvertisedStartTime.AsTime()) {
		r.Status = racing.Status_OPEN
	} else {
		r.Status = racing.Status_CLOSED
	}

	return r
}
