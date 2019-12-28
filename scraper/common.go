package scraper

import (
	"github.com/jmoiron/sqlx"
)

type scraper interface {
	Run(db *sqlx.DB, die <-chan struct{}) error
	Name() string
}

func Run(db *sqlx.DB, workers int, die <-chan struct{}) {
	sc := NewIpScraper(db, workers, die)
	sc.Run(db, die)
}
