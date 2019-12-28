package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/loicnestler/greeb/scraper"
)

var (
	verbose  = kingpin.Flag("verbose", "Verbose logging").Short('v').Bool()
	sqlite   = kingpin.Flag("sqlite", "").String()
	postgres = kingpin.Flag("postgres", "").String()
	scrape   = kingpin.Flag("scrape", "").Bool()
	workers  = kingpin.Flag("workers", "").Required().Int()
)

func main() {
	kingpin.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	var (
		db  *sqlx.DB
		err error
	)

	log.Info(sqlite)
	if *sqlite != "" {
		log.Info("Connecting to database")
		db, err = sqlx.Connect("sqlite3", *sqlite)
	} else if *postgres != "" {
		log.Info("Connecting to database")
		db, err = sqlx.Connect("postgres", *postgres)
	} else {
		log.Fatal("No database given")
	}

	if err != nil {
		log.WithField("err", err).Fatal("Error during datbase connection")
	}
	defer db.Close()

	_, err = db.Exec(schema)
	if err != nil {
		if err.Error() != "table addresses already exists" {
			log.Fatalf("%+v\n", err)
		}
	}

	if *scrape {
		log.Info("Starting scraping component")
		die := make(chan struct{})
		scraper.Run(db, *workers, die)
	}
}
