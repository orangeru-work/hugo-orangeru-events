package calendar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCategoryForPartStat(t *testing.T) {
	cfg := SyncConfig{
		GoingCategory:      "going-cat",
		InterestedCategory: "interested-cat",
		DefaultCategory:    "default-cat",
	}

	if got := categoryForPartStat("ACCEPTED", cfg); got != "going-cat" {
		t.Fatalf("ACCEPTED category = %q, want %q", got, "going-cat")
	}
	if got := categoryForPartStat("TENTATIVE", cfg); got != "interested-cat" {
		t.Fatalf("TENTATIVE category = %q, want %q", got, "interested-cat")
	}
	if got := categoryForPartStat("DECLINED", cfg); got != "default-cat" {
		t.Fatalf("DECLINED category = %q, want %q", got, "default-cat")
	}
}

func TestLoadCancelledEventIDs(t *testing.T) {
	dir := t.TempDir()
	cancelledPath := filepath.Join(dir, "cancelled-events.txt")

	content := strings.Join([]string{
		"# comment",
		"",
		"e100@facebook.com",
		"e200 trailing-text",
		"e300@facebook.com   note",
	}, "\n")
	if err := os.WriteFile(cancelledPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write cancelled file: %v", err)
	}

	cancelled, err := loadCancelledEventIDs(cancelledPath)
	if err != nil {
		t.Fatalf("load cancelled ids: %v", err)
	}

	for _, want := range []string{"e100", "e200", "e300"} {
		if _, ok := cancelled[want]; !ok {
			t.Fatalf("missing cancelled uid %q", want)
		}
	}
	if len(cancelled) != 3 {
		t.Fatalf("cancelled count = %d, want 3", len(cancelled))
	}
}

func TestRemoveGeneratedEventFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "e-old.md"), "old")
	writeFile(t, filepath.Join(dir, "e-keep.txt"), "txt")
	writeFile(t, filepath.Join(dir, "manual.md"), "manual")

	if err := removeGeneratedEventFiles(dir, "e-"); err != nil {
		t.Fatalf("remove generated files: %v", err)
	}

	assertExists(t, filepath.Join(dir, "manual.md"))
	assertExists(t, filepath.Join(dir, "e-keep.txt"))
	assertNotExists(t, filepath.Join(dir, "e-old.md"))
}

func TestSyncFacebookEvents(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}

	writeFile(t, filepath.Join(outputDir, "e-stale.md"), "stale generated")
	writeFile(t, filepath.Join(outputDir, "manual.md"), "manual event")

	cancelledPath := filepath.Join(dir, "cancelled-events.txt")
	writeFile(t, cancelledPath, "e-interested@facebook.com\n")

	icsPath := filepath.Join(dir, "events.ics")
	now := time.Now().UTC()
	goingStart := now.Add(24 * time.Hour)
	interestedStart := now.Add(48 * time.Hour)
	excludedStart := now.Add(72 * time.Hour)
	writeFile(t, icsPath, buildICS([]icsEvent{
		{
			UID:         "e-going@facebook.com",
			Summary:     "! Going Event",
			Organizer:   "Trail Club",
			PartStat:    "ACCEPTED",
			Start:       goingStart,
			End:         goingStart.Add(90 * time.Minute),
			Created:     now,
			URL:         "https://example.org/going",
			Location:    "Trailhead",
			Description: "Bring water",
		},
		{
			UID:         "e-interested@facebook.com",
			Summary:     "* Interested Event",
			Organizer:   "Trail Club",
			PartStat:    "TENTATIVE",
			Start:       interestedStart,
			End:         interestedStart.Add(90 * time.Minute),
			Created:     now,
			URL:         "https://example.org/interested",
			Location:    "Park",
			Description: "Bring snacks",
		},
		{
			UID:         "e-excluded@facebook.com",
			Summary:     "Excluded Event",
			Organizer:   "Skip Org",
			PartStat:    "ACCEPTED",
			Start:       excludedStart,
			End:         excludedStart.Add(90 * time.Minute),
			Created:     now,
			URL:         "https://example.org/excluded",
			Location:    "Downtown",
			Description: "Skip me",
		},
	}))

	err := SyncFacebookEvents(SyncConfig{
		ICSPath:             icsPath,
		OutputDir:           outputDir,
		CancelledEventsPath: cancelledPath,
		CleanupPrefix:       "e-",
		ExcludedOrganizers: map[string]struct{}{
			"Skip Org": {},
		},
		GoingCategory:      "going-cat",
		InterestedCategory: "interested-cat",
		DefaultCategory:    "default-cat",
	})
	if err != nil {
		t.Fatalf("sync facebook events: %v", err)
	}

	assertNotExists(t, filepath.Join(outputDir, "e-stale.md"))
	assertExists(t, filepath.Join(outputDir, "manual.md"))
	assertExists(t, filepath.Join(outputDir, "e-going.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-interested.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-excluded.md"))

	content, err := os.ReadFile(filepath.Join(outputDir, "e-going.md"))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `category: going-cat`) {
		t.Fatalf("generated file missing configured going category")
	}
	if !strings.Contains(text, `title: "Going Event"`) {
		t.Fatalf("generated file missing cleaned title")
	}
}

type icsEvent struct {
	UID         string
	Summary     string
	Description string
	URL         string
	Start       time.Time
	End         time.Time
	Created     time.Time
	Location    string
	Organizer   string
	PartStat    string
}

func buildICS(events []icsEvent) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\n")
	b.WriteString("VERSION:2.0\n")
	b.WriteString("PRODID:-//hugo-orangeru-events//tests//EN\n")
	for _, e := range events {
		b.WriteString("BEGIN:VEVENT\n")
		b.WriteString(fmt.Sprintf("UID:%s\n", e.UID))
		b.WriteString("CLASS:PUBLIC\n")
		b.WriteString(fmt.Sprintf("SUMMARY:%s\n", e.Summary))
		b.WriteString(fmt.Sprintf("DESCRIPTION:%s\n", e.Description))
		b.WriteString(fmt.Sprintf("URL:%s\n", e.URL))
		b.WriteString(fmt.Sprintf("DTSTART:%s\n", e.Start.UTC().Format("20060102T150405Z")))
		b.WriteString(fmt.Sprintf("DTEND:%s\n", e.End.UTC().Format("20060102T150405Z")))
		b.WriteString(fmt.Sprintf("DTSTAMP:%s\n", e.Created.UTC().Format("20060102T150405Z")))
		b.WriteString(fmt.Sprintf("CREATED:%s\n", e.Created.UTC().Format("20060102T150405Z")))
		b.WriteString(fmt.Sprintf("LAST-MODIFIED:%s\n", e.Created.UTC().Format("20060102T150405Z")))
		b.WriteString(fmt.Sprintf("LOCATION:%s\n", e.Location))
		b.WriteString(fmt.Sprintf("ORGANIZER;CN=%s:mailto:test@example.com\n", e.Organizer))
		b.WriteString(fmt.Sprintf("ATTENDEE;PARTSTAT=%s:mailto:member@example.com\n", e.PartStat))
		b.WriteString("END:VEVENT\n")
	}
	b.WriteString("END:VCALENDAR\n")
	return b.String()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s not to exist, got err=%v", path, err)
	}
}
