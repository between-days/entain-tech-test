package service

import (
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// unit test as it's a simple function with no integration concern
// integration test at races_test.go already gives us basic integration checks
// only going to bother writing tests for new functionality
func TestEnrichRace_Status(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		race     *racing.Race
		expected racing.Status
	}{
		{
			name: "future race is OPEN",
			race: &racing.Race{
				AdvertisedStartTime: mustProtoTime(now.Add(2 * time.Hour)),
			},
			expected: racing.Status_OPEN,
		},
		{
			name: "past race is CLOSED",
			race: &racing.Race{
				AdvertisedStartTime: mustProtoTime(now.Add(-2 * time.Hour)),
			},
			expected: racing.Status_CLOSED,
		},
		{
			name: "exact time is CLOSED",
			race: &racing.Race{
				AdvertisedStartTime: mustProtoTime(now),
			},
			expected: racing.Status_CLOSED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enrichRace(tt.race, now)

			if got.Status != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					got.Status,
				)
			}
		})
	}
}

//
// helpersa
//

func mustProtoTime(t time.Time) *timestamppb.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return ts
}
