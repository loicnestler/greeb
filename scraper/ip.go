package scraper

import (
	"errors"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"net/http"
	// "net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ipScraper struct {
	addrs   chan string
	die     <-chan struct{}
	db      *sqlx.DB
	workers int
}

func NewIpScraper(db *sqlx.DB, workers int, die <-chan struct{}) *ipScraper {
	return &ipScraper{
		addrs:   make(chan string),
		die:     die,
		db:      db,
		workers: workers,
	}
}

func (s *ipScraper) Run(db *sqlx.DB, die <-chan struct{}) {
	go s.genIPs()

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go s.checkPHP(&wg)
	}

	wg.Wait()
}

func (s *ipScraper) Name() string {
	return "IP Scraper"
}

func (s *ipScraper) genIPs() {
	go func() {
		time.Sleep(time.Duration(rand.Int31n(1000*10)) * time.Millisecond)
		s.addrs <- "103.10.56.21"
		s.addrs <- "128.177.134.252"
	}()

	for i1 := 100; i1 < 255; i1++ {
		for i2 := 0; i2 < 255; i2++ {
			for i3 := 0; i3 < 255; i3++ {
				for i4 := 1; i4 < 255; i4++ {
					ip := strconv.Itoa(i1) + "." + strconv.Itoa(i2) + "." + strconv.Itoa(i3) + "." + strconv.Itoa(i4)

					select {
					case s.addrs <- ip:
					case <-s.die:
						return
					}
				}
			}
		}
	}
}

func (s *ipScraper) checkPHP(wg *sync.WaitGroup) {
	defer wg.Done()

	// client := http.DefaultClient
	// client.Timeout = time.Duration(20 * time.Second)

	for {
		var ip string

		select {
		case <-s.die:
			return
		case i := <-s.addrs:
			ip = i
		}

		if ip == "" {
			return
		}

		log.WithField("ip", ip).Debug("Checking for PHP5")

		resp, err := http.Get("http://" + ip)

		if err != nil {
			var (
				opError *net.OpError
				// urlError *url.Error
			)

			log.Errorf("%+v %T", err, err)
			if errors.As(err, &opError) {
				log.WithError(opError).Trace("Timeout, I guess?")
				// } else if errors.As(err, &urlError) {
				// 	log.WithError(urlError).Trace("Timeout, I guess?")
			} else {
				log.Errorf("%+v %T", err, err)
			}

			continue
		}
		resp.Body.Close()

		xPoweredBy := resp.Header.Get("X-Powered-By")
		server := resp.Header.Get("Server")

		if strings.Contains(xPoweredBy, "PHP/5") || strings.Contains(xPoweredBy, "php/5") || strings.Contains(server, "PHP/5") || strings.Contains(server, "php/5") {
			logger := log.WithField("io", ip)
			logger.Info("Found server with PHP 5")

			tx, err := s.db.Begin()
			if err != nil {
				logger.WithError(err).Error("Worker thread failed to access database")
				tx.Rollback()
				return
			}

			_, err = tx.Exec("INSERT INTO addresses (address, analyzed) VALUES ($1, FALSE)", ip)
			if err != nil {
				if err.Error() != "UNIQUE constraint failed: addresses.address" {
					logger.WithError(err).Error("Worker failed to insert into database")
				}

				tx.Rollback()
				return
			}

			err = tx.Commit()
			if err != nil {
				logger.WithError(err).Error("Worker failed to commit database")
				tx.Rollback()
				return
			}
		}
	}
}
