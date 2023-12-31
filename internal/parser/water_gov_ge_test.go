package parser

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/doesnotcommit/outage_monitor/internal/outage"
	"github.com/stretchr/testify/assert"
)

func Test_ParseMapMarkers(t *testing.T) {
	rawMapMarkers, err := os.ReadFile("./fixtures/map.html")
	if err != nil {
		t.Fatal(err)
	}
	w, err := NewWaterGovGe(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	m, err := w.parseMapMarkers(ctx, rawMapMarkers)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) == 0 {
		t.Fatal("map parser is broken")
	}
}

func Test_ParseProblem(t *testing.T) {
	rawProblem, err := os.ReadFile("./fixtures/problem.html")
	if err != nil {
		t.Fatal(err)
	}
	w, err := NewWaterGovGe(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	location := outage.Location{
		Id:       "1",
		TitleLat: "whatever",
		Lat:      "30.31",
		Lng:      "42.42",
	}
	p, err := w.parseProblem(ctx, location, rawProblem)
	if err != nil {
		t.Fatal(err)
	}
	wantP := outage.WaterGovGe{
		Location:          location,
		Start:             time.Date(2023, 9, 8, 19, 20, 0, 0, w.location),
		End:               time.Date(2023, 9, 11, 19, 20, 0, 0, w.location),
		AffectedCustomers: 186,
		AddressesGe: []string{
			"ოზურგეთი ე.თაყაიშვილის I შეს.",
			"ოზურგეთი ე.თაყაიშვილის II შეს.",
			"ოზურგეთი ე.თაყაიშვილის III შეს.",
			"ოზურგეთი ე.თაყაიშვილის IV შეს.",
			"ოზურგეთი ე.თაყაიშვილის V შეს.",
			"ოზურგეთი ე.თაყაიშვილის I ჩიხი",
			"ოზურგეთი ე.თაყაიშვილის II ჩიხი",
			"ოზურგეთი ე.თაყაიშვილის IV ჩიხი",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 15",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 5",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 32",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 57",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 20",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 42",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 17",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 36",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 31",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 41",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 12",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 85",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 39ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 79",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 31ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 6",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 62",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 48",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 25",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 54",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 22",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 93",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 35",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 46",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 10",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 71ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 21ბ",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 44",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 2",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 58",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 9",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 52",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 34",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 91",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 55",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 19",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 11",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 39",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 26",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 37",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 21",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 43",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 30",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 24",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 99",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 47",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 115",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 1",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 64",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 13",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 53",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 101",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 2ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 7",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 32ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 60",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 107",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 97",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 40",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 66",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 83",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 16",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 29",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 103",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 8",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 3",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 38",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 109",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 8ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 35ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 41ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 18ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 33",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 75",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 105",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 111",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 73",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 71",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 77",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 45",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 28",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 87",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 4ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 27",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 42ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 14",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 1",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 4",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 56",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 89",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 40ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 71ბ",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 42ბ",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 29ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 5ა",
			"ოზურგეთი ე.თაყაიშვილის ქ. N 32ბ",
		},
	}
	assert.Equal(t, wantP, p)
}
