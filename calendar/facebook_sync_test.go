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
	writeFile(t, icsPath, facebookFixture(time.Now().UTC()))

	err := SyncFacebookEvents(SyncConfig{
		ICSPath:             icsPath,
		OutputDir:           outputDir,
		CancelledEventsPath: cancelledPath,
		CleanupPrefix:       "e-",
		GoingCategory:       "going-cat",
		InterestedCategory:  "interested-cat",
		DefaultCategory:     "default-cat",
	})
	if err != nil {
		t.Fatalf("sync facebook events: %v", err)
	}

	assertNotExists(t, filepath.Join(outputDir, "e-stale.md"))
	assertExists(t, filepath.Join(outputDir, "manual.md"))
	assertExists(t, filepath.Join(outputDir, "e-going.md"))
	assertExists(t, filepath.Join(outputDir, "e-attendee-going.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-interested.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-declined.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-private.md"))

	content, err := os.ReadFile(filepath.Join(outputDir, "e-going.md"))
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "\nevent:\n") {
		t.Fatalf("generated file should use contained event front matter")
	}
	if !strings.Contains(text, `category: going-cat`) {
		t.Fatalf("generated file missing configured going category")
	}
	if !strings.Contains(text, `title: "Going Event"`) {
		t.Fatalf("generated file missing cleaned title")
	}
	if strings.Contains(text, `feature-img:`) {
		t.Fatalf("generated file should not set feature-img")
	}

	attendeeContent, err := os.ReadFile(filepath.Join(outputDir, "e-attendee-going.md"))
	if err != nil {
		t.Fatalf("read attendee-status file: %v", err)
	}
	if !strings.Contains(string(attendeeContent), `category: going-cat`) {
		t.Fatalf("attendee-derived status should map to going category")
	}
}

func TestSyncFacebookEventsIncludePrivate(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "events")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}

	icsPath := filepath.Join(dir, "events.ics")
	writeFile(t, icsPath, facebookFixture(time.Now().UTC()))

	err := SyncFacebookEvents(SyncConfig{
		ICSPath:         icsPath,
		OutputDir:       outputDir,
		CleanupPrefix:   "e-",
		IncludePrivate:  true,
		GoingCategory:   "going-cat",
		DefaultCategory: "default-cat",
	})
	if err != nil {
		t.Fatalf("sync with private events included: %v", err)
	}

	assertExists(t, filepath.Join(outputDir, "e-private.md"))
	assertNotExists(t, filepath.Join(outputDir, "e-declined.md"))
	content, err := os.ReadFile(filepath.Join(outputDir, "e-private.md"))
	if err != nil {
		t.Fatalf("read private event: %v", err)
	}
	if !strings.Contains(string(content), `category: going-cat`) {
		t.Fatalf("private event category should follow partstat mapping")
	}
}

func facebookFixture(base time.Time) string {
	created := base.Format("20060102T150405Z")

	goingStart := base.Add(24 * time.Hour).Format("20060102T150405Z")
	goingEnd := base.Add(25 * time.Hour).Format("20060102T150405Z")
	interestedStart := base.Add(48 * time.Hour).Format("20060102T150405Z")
	interestedEnd := base.Add(49 * time.Hour).Format("20060102T150405Z")
	attendeeStart := base.Add(72 * time.Hour).Format("20060102T150405Z")
	attendeeEnd := base.Add(73 * time.Hour).Format("20060102T150405Z")
	privateStart := base.Add(96 * time.Hour).Format("20060102T150405Z")
	privateEnd := base.Add(97 * time.Hour).Format("20060102T150405Z")
	declinedStart := base.Add(120 * time.Hour).Format("20060102T150405Z")
	declinedEnd := base.Add(121 * time.Hour).Format("20060102T150405Z")

	// The shape intentionally mirrors Facebook ICS fields, folding, and escaping.
	return fmt.Sprintf(`BEGIN:VCALENDAR
PRODID:-//Facebook//NONSGML Facebook Events V1.0//EN
X-WR-CALNAME:Fixture Facebook Events
X-PUBLISHED-TTL:PT12H
X-ORIGINAL-URL:/events/
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
BEGIN:VEVENT
DTSTAMP:%[1]s
LAST-MODIFIED:%[1]s
CREATED:%[1]s
SEQUENCE:3476742
ORGANIZER;CN=Fixture Organizer:MAILTO:noreply@facebookmail.com
DTSTART:%[2]s
DTEND:%[3]s
UID:e-going@facebook.com
SUMMARY:! Going Event
LOCATION:3181 Willow Creek Rd\, Prescott\, AZ
  86301-6848\, United States
URL:https://www.facebook.com/events/1111/?event_time_id=2222
DESCRIPTION:Join us for a 5k at the Jan Alfano
  parkrun each Saturday morning! We
  start at 7:30am.\n\nhttps://www.fac
 ebook.com/events/1111/?event_time_
 id=2222
CLASS:PUBLIC
STATUS:CONFIRMED
PARTSTAT:ACCEPTED
END:VEVENT
BEGIN:VEVENT
DTSTAMP:%[1]s
LAST-MODIFIED:%[1]s
CREATED:%[1]s
SEQUENCE:3476742
ORGANIZER;CN=Fixture Organizer:MAILTO:noreply@facebookmail.com
DTSTART:%[4]s
DTEND:%[5]s
UID:e-interested@facebook.com
SUMMARY:* Interested Event
LOCATION:City Park\, Prescott\, AZ
URL:https://www.facebook.com/events/3333/?event_time_id=4444
DESCRIPTION:Bring snacks and water
CLASS:PUBLIC
STATUS:CONFIRMED
PARTSTAT:TENTATIVE
END:VEVENT
BEGIN:VEVENT
DTSTAMP:%[1]s
LAST-MODIFIED:%[1]s
CREATED:%[1]s
SEQUENCE:3476742
ORGANIZER;CN=Fixture Organizer:MAILTO:noreply@facebookmail.com
DTSTART:%[6]s
DTEND:%[7]s
UID:e-attendee-going@facebook.com
SUMMARY:Attendee Status Event
LOCATION:Trailhead\, Prescott\, AZ
URL:https://www.facebook.com/events/5555/?event_time_id=6666
DESCRIPTION:Status only from attendee
CLASS:PUBLIC
STATUS:CONFIRMED
ATTENDEE;CUTYPE=INDIVIDUAL;ROLE=REQ-PARTICIPANT;PARTSTAT=ACCEPTED;CN=Fixture User:MAILTO:fixture@example.com
END:VEVENT
BEGIN:VEVENT
DTSTAMP:%[1]s
LAST-MODIFIED:%[1]s
CREATED:%[1]s
SEQUENCE:3476742
ORGANIZER;CN=Fixture Organizer:MAILTO:noreply@facebookmail.com
DTSTART:%[8]s
DTEND:%[9]s
UID:e-private@facebook.com
SUMMARY:Private Event
LOCATION:Private\, Prescott\, AZ
URL:https://www.facebook.com/events/7777/?event_time_id=8888
DESCRIPTION:Should be skipped because it is private
CLASS:PRIVATE
STATUS:CONFIRMED
PARTSTAT:ACCEPTED
END:VEVENT
BEGIN:VEVENT
DTSTAMP:%[1]s
LAST-MODIFIED:%[1]s
CREATED:%[1]s
SEQUENCE:3476742
ORGANIZER;CN=Fixture Organizer:MAILTO:noreply@facebookmail.com
DTSTART:%[10]s
DTEND:%[11]s
UID:e-declined@facebook.com
SUMMARY:Declined Event
LOCATION:Declined\, Prescott\, AZ
URL:https://www.facebook.com/events/9999/?event_time_id=1111
DESCRIPTION:Should be skipped because the user declined
CLASS:PUBLIC
STATUS:CONFIRMED
PARTSTAT:DECLINED
END:VEVENT
END:VCALENDAR
`, created, goingStart, goingEnd, interestedStart, interestedEnd, attendeeStart, attendeeEnd, privateStart, privateEnd, declinedStart, declinedEnd)
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
