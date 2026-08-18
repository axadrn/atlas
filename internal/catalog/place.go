package catalog

type PlaceType string

const (
	PlaceTypeCountry      PlaceType = "country"
	PlaceTypeRegion       PlaceType = "region"
	PlaceTypeLocality     PlaceType = "locality"
	PlaceTypeNeighborhood PlaceType = "neighborhood"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Bounds struct {
	West  float64 `json:"west"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	North float64 `json:"north"`
}

type PlaceSummary struct {
	ID          string       `json:"id"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Type        PlaceType    `json:"type"`
	ParentID    string       `json:"parent_id,omitempty"`
	CountryCode string       `json:"country_code"`
	Destination bool         `json:"is_destination"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
	Bounds      *Bounds      `json:"bounds,omitempty"`
	Timezones   []string     `json:"timezones,omitempty"`
	Population  *int64       `json:"population,omitempty"`
	// CuratedRank orders destinations by nomad relevance, lower is better.
	// NULL means unranked; see migration 0004.
	CuratedRank *int64 `json:"curated_rank,omitempty"`
}

type SourceAttribution struct {
	Name         string
	HomepageURL  string
	LicenseName  string
	LicenseURL   string
	ExternalID   string
	RecordURL    string
	Contribution string
	RetrievedAt  string
}
