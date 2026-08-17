package placesearch_test

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"atlas/components/placesearch"
)

func TestPlaceSearchComposesComboboxPrimitives(t *testing.T) {
	html, err := templ.ToGoHTML(context.Background(), placesearch.PlaceSearch())
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		`data-slot="input-group"`,
		`data-tui-combobox-input`,
		`data-tui-combobox-content`,
		`data-tui-combobox-list`,
	} {
		if !strings.Contains(string(html), marker) {
			t.Fatalf("place search is missing combobox primitive %q", marker)
		}
	}
}
