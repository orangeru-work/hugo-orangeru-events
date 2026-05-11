package calendar

import (
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
	writeFile(t, icsPath, facebookFixture(t, time.Now().UTC()))

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
	writeFile(t, icsPath, facebookFixture(t, time.Now().UTC()))

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

func facebookFixture(t *testing.T, base time.Time) string {
	t.Helper()
	rawFixture, err := os.ReadFile(filepath.Join("testdata", "facebook_fixture.ics"))
	if err != nil {
		t.Fatalf("read facebook fixture: %v", err)
	}

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

	replacer := strings.NewReplacer(
		"{{CREATED}}", created,
		"{{GOING_START}}", goingStart,
		"{{GOING_END}}", goingEnd,
		"{{INTERESTED_START}}", interestedStart,
		"{{INTERESTED_END}}", interestedEnd,
		"{{ATTENDEE_START}}", attendeeStart,
		"{{ATTENDEE_END}}", attendeeEnd,
		"{{PRIVATE_START}}", privateStart,
		"{{PRIVATE_END}}", privateEnd,
		"{{DECLINED_START}}", declinedStart,
		"{{DECLINED_END}}", declinedEnd,
	)

	return replacer.Replace(string(rawFixture))
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
