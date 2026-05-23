package eventsmodule_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleSiteBuildIncludesNestedEventsAndICSFeed(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "public")
	cmd := exec.Command("go", "run", "github.com/gohugoio/hugo@v0.150.0", "--source", "exampleSite", "--destination", outDir)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build example site: %v\n%s", err, output)
	}

	feed, err := os.ReadFile(filepath.Join(outDir, "events", "index.ics"))
	if err != nil {
		t.Fatalf("read generated feed: %v", err)
	}

	feedText := string(feed)
	if !strings.Contains(feedText, "SUMMARY:Demo Event Nested") {
		t.Fatalf("generated feed should include nested event, got:\n%s", feedText)
	}
	if strings.Contains(feedText, "X-PUBLISHED-TTL:PT1HBEGIN:VEVENT") {
		t.Fatalf("generated feed should keep calendar headers and VEVENT lines separated, got:\n%s", feedText)
	}

	nestedPage, err := os.ReadFile(filepath.Join(outDir, "events", "2026", "06", "demo-event-nested", "index.html"))
	if err != nil {
		t.Fatalf("read nested event page: %v", err)
	}

	nestedText := string(nestedPage)
	if !strings.Contains(nestedText, "Demo Event Nested") {
		t.Fatalf("nested event page missing title, got:\n%s", nestedText)
	}
	if !strings.Contains(nestedText, "Add to Calendar") {
		t.Fatalf("nested event page missing calendar link, got:\n%s", nestedText)
	}
	for _, want := range []string{
		`<script type="application/ld+json">{"@context":"https://schema.org","@type":"Event"`,
		`"startDate":"2026-06-18T18:30:00-07:00"`,
		`"endDate":"2026-06-18T20:00:00-07:00"`,
		`"location":{"@type":"Place","name":"Nested Folder Venue"}`,
		`"description":"This sample event lives in a nested year/month folder to verify recursive event discovery."`,
	} {
		if !strings.Contains(nestedText, want) {
			t.Fatalf("nested event page missing %q, got:\n%s", want, nestedText)
		}
	}
	if strings.Contains(nestedText, `<script type="application/ld+json">"{`) {
		t.Fatalf("nested event schema should render raw JSON, got:\n%s", nestedText)
	}

	disabledPage, err := os.ReadFile(filepath.Join(outDir, "events", "demo-event-schema-disabled", "index.html"))
	if err != nil {
		t.Fatalf("read disabled-schema event page: %v", err)
	}

	disabledText := string(disabledPage)
	if !strings.Contains(disabledText, "Demo Event Schema Disabled") {
		t.Fatalf("disabled-schema event page missing title, got:\n%s", disabledText)
	}
	if strings.Contains(disabledText, `application/ld+json`) {
		t.Fatalf("disabled-schema event page should omit Event JSON-LD, got:\n%s", disabledText)
	}

	calendarPage, err := os.ReadFile(filepath.Join(outDir, "events", "index.html"))
	if err != nil {
		t.Fatalf("read calendar page: %v", err)
	}

	calendarText := string(calendarPage)
	if strings.Contains(calendarText, "Select an event to preview details and links.") {
		t.Fatalf("calendar page should not render the old inline details prompt, got:\n%s", calendarText)
	}
	if !strings.Contains(calendarText, "events-calendar-hovercard") {
		t.Fatalf("calendar page missing hovercard markup, got:\n%s", calendarText)
	}
}
