package catalog

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type PlaceKind string

const (
	PlaceKindWorld     PlaceKind = "world"
	PlaceKindContinent PlaceKind = "continent"
	PlaceKindCountry   PlaceKind = "country"
	PlaceKindTerritory PlaceKind = "territory"
	PlaceKindRegion    PlaceKind = "region"
	PlaceKindCity      PlaceKind = "city"
	PlaceKindMetro     PlaceKind = "metro"
)

type PlaceStatus string

const (
	PlaceStatusActive   PlaceStatus = "active"
	PlaceStatusHistoric PlaceStatus = "historic"
	PlaceStatusDisputed PlaceStatus = "disputed"
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Place struct {
	ID          string
	Slug        string
	Kind        PlaceKind
	Status      PlaceStatus
	ParentID    string
	CountryCode string
	Coordinates *Coordinates
	Timezone    string
	Population  *int64
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (p Place) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("place ID is required")
	}
	if !slugPattern.MatchString(p.Slug) {
		return errors.New("place slug must contain lowercase letters, numbers and single hyphens")
	}
	if !p.Kind.valid() {
		return fmt.Errorf("invalid place kind %q", p.Kind)
	}
	if !p.Status.valid() {
		return fmt.Errorf("invalid place status %q", p.Status)
	}
	if p.ParentID == p.ID {
		return errors.New("place cannot be its own parent")
	}
	if p.CountryCode != "" {
		if len(p.CountryCode) != 2 || p.CountryCode != strings.ToUpper(p.CountryCode) {
			return errors.New("country code must be a two-letter uppercase code")
		}
	}
	if p.Coordinates != nil {
		if p.Coordinates.Latitude < -90 || p.Coordinates.Latitude > 90 {
			return errors.New("latitude must be between -90 and 90")
		}
		if p.Coordinates.Longitude < -180 || p.Coordinates.Longitude > 180 {
			return errors.New("longitude must be between -180 and 180")
		}
	}
	if p.Population != nil && *p.Population < 0 {
		return errors.New("population cannot be negative")
	}
	return nil
}

func (k PlaceKind) valid() bool {
	switch k {
	case PlaceKindWorld, PlaceKindContinent, PlaceKindCountry, PlaceKindTerritory,
		PlaceKindRegion, PlaceKindCity, PlaceKindMetro:
		return true
	default:
		return false
	}
}

func (s PlaceStatus) valid() bool {
	switch s {
	case PlaceStatusActive, PlaceStatusHistoric, PlaceStatusDisputed:
		return true
	default:
		return false
	}
}
