// Command import refreshes GeoNames-sourced facts in the Atlas database
// from a standard GeoNames dump file (for example cities500.txt from
// https://download.geonames.org/export/dump/).
//
// The run is update-only: it never adds or removes places and never touches
// curated fields. Use -dry-run to see what would happen first.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"atlas/internal/database"
	"atlas/internal/geonames"
)

func main() {
	dbPath := flag.String("db", defaultDatabasePath(), "path to the Atlas SQLite database")
	dumpPath := flag.String("file", "", "path to a GeoNames dump file (tab separated, 19 columns)")
	dryRun := flag.Bool("dry-run", false, "report what would change without writing")
	flag.Parse()

	if *dumpPath == "" {
		fmt.Fprintln(os.Stderr, "usage: import -file cities500.txt [-db atlas.db] [-dry-run]")
		os.Exit(2)
	}

	ctx := context.Background()
	db, err := database.Open(ctx, *dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	dump, err := os.Open(*dumpPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer dump.Close()

	stats, err := geonames.Refresh(ctx, db, dump, time.Now(), *dryRun)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mode := "updated"
	if *dryRun {
		mode = "would update"
	}
	fmt.Printf("scanned %d dump rows, %s %d known places\n", stats.Scanned, mode, stats.Matched)
}

func defaultDatabasePath() string {
	if path := os.Getenv("DATABASE_PATH"); path != "" {
		return path
	}
	return "atlas.db"
}
