package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/orangeru-work/hugo-orangeru-events/calendar"
)

type repeatedFlag []string

func (f *repeatedFlag) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *repeatedFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func main() {
	var (
		icsPath            string
		outputDir          string
		cancelledPath      string
		cleanupPrefix      string
		goingCategory      string
		interestedCategory string
		defaultCategory    string
		excludes           repeatedFlag
	)

	flag.StringVar(&icsPath, "ics", "", "Path to source ICS file")
	flag.StringVar(&outputDir, "output", "", "Path to Hugo events content directory")
	flag.StringVar(&cancelledPath, "cancelled", "", "Path to cancelled-events file (optional)")
	flag.StringVar(&cleanupPrefix, "cleanup-prefix", "e", "Delete existing generated markdown files matching this prefix")
	flag.StringVar(&goingCategory, "category-going", "club", "Category to assign for GOING/ACCEPTED events")
	flag.StringVar(&interestedCategory, "category-interested", "other", "Category to assign for INTERESTED/TENTATIVE events")
	flag.StringVar(&defaultCategory, "category-default", "other", "Category to assign for other participant states")
	flag.Var(&excludes, "exclude-organizer", "Organizer display name to exclude (repeatable)")
	flag.Parse()

	if icsPath == "" || outputDir == "" {
		flag.Usage()
		os.Exit(2)
	}

	excludedOrganizers := make(map[string]struct{}, len(excludes))
	for _, name := range excludes {
		excludedOrganizers[name] = struct{}{}
	}

	cfg := calendar.SyncConfig{
		ICSPath:             icsPath,
		OutputDir:           outputDir,
		CancelledEventsPath: cancelledPath,
		CleanupPrefix:       cleanupPrefix,
		ExcludedOrganizers:  excludedOrganizers,
		GoingCategory:       goingCategory,
		InterestedCategory:  interestedCategory,
		DefaultCategory:     defaultCategory,
	}
	if err := calendar.SyncFacebookEvents(cfg); err != nil {
		log.Fatal(err)
	}
}
