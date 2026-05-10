package calendar

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bakerag1/gocal"
)

const outputFmt = `---
title: "%v"
date: %v
startdate: %v
enddate: %v
category: %v
external_url: %v
layout: %v
location: %v
feature-img: "assets/img/big-trail.jpg"
outputs:
  - html
  - calendar
ICSDescription: |+2
  %s
---

%v
`

// SyncConfig controls Facebook ICS -> Hugo event page generation.
type SyncConfig struct {
	ICSPath             string
	OutputDir           string
	CancelledEventsPath string
	CleanupPrefix       string
	ExcludedOrganizers  map[string]struct{}
	IncludePrivate      bool
	GoingCategory       string
	InterestedCategory  string
	DefaultCategory     string
}

type event struct {
	URI            string
	ICSDescription string
	Description    string
	Summary        string
	Start          string
	End            string
	Location       string
	UID            string
	Created        string
	Category       string
}

// SyncFacebookEvents renders markdown event pages from an ICS feed.
func SyncFacebookEvents(cfg SyncConfig) error {
	if cfg.ICSPath == "" {
		return fmt.Errorf("ics path is required")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if cfg.CleanupPrefix == "" {
		cfg.CleanupPrefix = "e"
	}
	if cfg.GoingCategory == "" {
		cfg.GoingCategory = "club"
	}
	if cfg.InterestedCategory == "" {
		cfg.InterestedCategory = "other"
	}
	if cfg.DefaultCategory == "" {
		cfg.DefaultCategory = cfg.InterestedCategory
	}

	cancelled, err := loadCancelledEventIDs(cfg.CancelledEventsPath)
	if err != nil {
		return err
	}

	if err := removeGeneratedEventFiles(cfg.OutputDir, cfg.CleanupPrefix); err != nil {
		return err
	}

	events, err := parseEvents(cfg)
	if err != nil {
		return err
	}

	for _, e := range events {
		if _, isCancelled := cancelled[e.UID]; isCancelled {
			continue
		}

		outputPath := filepath.Join(cfg.OutputDir, e.UID+".md")
		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}

		if _, err := f.Write([]byte(fmt.Sprintf(outputFmt,
			e.Summary,
			e.Created,
			e.Start,
			e.End,
			e.Category,
			e.URI,
			"post",
			e.Location,
			e.ICSDescription,
			e.Description,
		))); err != nil {
			_ = f.Close()
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

func removeGeneratedEventFiles(outputDir, cleanupPrefix string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, cleanupPrefix) || !strings.HasSuffix(name, ".md") {
			continue
		}

		if err := os.Remove(filepath.Join(outputDir, name)); err != nil {
			return err
		}
	}

	return nil
}

func loadCancelledEventIDs(path string) (map[string]struct{}, error) {
	cancelled := make(map[string]struct{})
	if path == "" {
		return cancelled, nil
	}

	fileContents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	for _, rawLine := range strings.Split(string(fileContents), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		uid := strings.Fields(line)[0]
		if at := strings.IndexByte(uid, '@'); at > 0 {
			uid = uid[:at]
		}
		cancelled[uid] = struct{}{}
	}

	return cancelled, nil
}

func parseEvents(cfg SyncConfig) ([]event, error) {
	f, err := os.Open(cfg.ICSPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	parser := gocal.NewParser(f)
	start, end := time.Now().Add(-48*time.Hour), time.Now().Add(120*30*24*time.Hour)
	parser.Start, parser.End = &start, &end
	parser.Strict.Mode = gocal.StrictModeFailAttribute
	parser.Parse()

	var events []event
	for _, e := range parser.Events {
		if !cfg.IncludePrivate && strings.EqualFold(strings.TrimSpace(e.Class), "PRIVATE") {
			continue
		}
		if _, excluded := cfg.ExcludedOrganizers[e.Organizer.Cn]; excluded {
			continue
		}

		uid := e.Uid
		if at := strings.IndexByte(uid, '@'); at > 0 {
			uid = uid[:at]
		}
		if uid == "" {
			continue
		}

		icsDescription := strings.ReplaceAll(e.Description, "\\n", "\n  ")
		description := strings.ReplaceAll(e.Description, "\\n", "<br>\n  ")
		description = strings.ReplaceAll(description, e.URL, "")
		description = strings.ReplaceAll(description, ":", "&#58;")
		description = strings.ReplaceAll(description, "\n\n", "<br>\n  ")

		urlExpr := regexp.MustCompile(`(http[s]?)&#58;(//[^ <)]*)`)
		description = urlExpr.ReplaceAllString(description, "[$1:$2]($1:$2)")

		coordExpr := regexp.MustCompile(`([0-9]{1,3}\.[0-9]*)° (N), ([0-9]{1,3}\.[0-9]*)(° W)`)
		description = coordExpr.ReplaceAllString(description, "[$1$2, $3$4](https://www.google.com/maps/place/$1,-$3)")
		icsDescription = coordExpr.ReplaceAllString(icsDescription, "$1° $2, $3$4 https://www.google.com/maps/place/$1,-$3")

		coordExpr2 := regexp.MustCompile(`([0-9]{1,3}\.[0-9]*), -([0-9]{1,3}\.[0-9]*)`)
		description = coordExpr2.ReplaceAllString(description, "[$1° N, $2° W](https://www.google.com/maps/place/$1,-$2)")
		icsDescription = coordExpr2.ReplaceAllString(icsDescription, "$1° N, $2° W https://www.google.com/maps/place/$1,-$2")

		summaryExpr := regexp.MustCompile(`^([^A-z0-9]*)(.*)`)
		summary := summaryExpr.ReplaceAllString(e.Summary, "$2")

		events = append(events, event{
			URI:            e.URL,
			ICSDescription: icsDescription,
			Description:    description,
			Summary:        summary,
			Start:          e.Start.Format("2006-01-02T15:04:00Z"),
			End:            e.End.Format("2006-01-02T15:04:00Z"),
			Location:       e.Location,
			UID:            uid,
			Created:        e.Created.Format("2006-01-02T15:04:00Z"),
			Category:       categoryForPartStat(resolvePartStat(e), cfg),
		})
	}

	sort.Slice(events, func(a, b int) bool {
		return strings.Compare(events[a].Start, events[b].Start) < 0
	})

	return events, nil
}

func categoryForPartStat(partStat string, cfg SyncConfig) string {
	switch strings.ToUpper(partStat) {
	case "ACCEPTED":
		return cfg.GoingCategory
	case "TENTATIVE":
		return cfg.InterestedCategory
	default:
		return cfg.DefaultCategory
	}
}

func resolvePartStat(e gocal.Event) string {
	if e.PartStat != "" {
		return e.PartStat
	}

	foundTentative := false
	for _, attendee := range e.Attendees {
		status := strings.ToUpper(attendee.Status)
		switch status {
		case "ACCEPTED":
			return status
		case "TENTATIVE":
			foundTentative = true
		case "":
			continue
		default:
			return status
		}
	}

	if foundTentative {
		return "TENTATIVE"
	}

	return ""
}
