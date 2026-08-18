package catalog

type PlaceKind string

const (
	PlaceKindCountry     PlaceKind = "country"
	PlaceKindDestination PlaceKind = "destination"
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
	Kind        PlaceKind    `json:"kind"`
	CountryCode string       `json:"country_code"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
	Bounds      *Bounds      `json:"bounds,omitempty"`
	Timezone    string       `json:"timezone,omitempty"`
	Population  *int64       `json:"population,omitempty"`
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
