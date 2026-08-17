package inputgroup_test

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"atlas/components/inputgroup"
)

func TestInputGroupClassesRemainSeparate(t *testing.T) {
	buttonHTML, err := templ.ToGoHTML(context.Background(), inputgroup.Button(inputgroup.ButtonProps{
		Size: inputgroup.ButtonSizeIconXs,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buttonHTML), "size-6") || strings.Contains(string(buttonHTML), "shadow-nonesize-6") {
		t.Fatalf("icon button classes are malformed: %s", buttonHTML)
	}

	addonHTML, err := templ.ToGoHTML(context.Background(), inputgroup.Addon(inputgroup.AddonProps{
		Align: inputgroup.AlignInlineEnd,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(addonHTML), "pr-2") || strings.Contains(string(addonHTML), "select-nonepr-2") {
		t.Fatalf("addon classes are malformed: %s", addonHTML)
	}
}
