package outage

import "time"

type Location struct {
	Id       string
	TitleGe  string
	TitleLat string
	Lat      string
	Lng      string
}

type WaterGovGe struct {
	Start             time.Time
	End               time.Time
	AffectedCustomers int
	Location          Location
	AddressesGe       []string
}
