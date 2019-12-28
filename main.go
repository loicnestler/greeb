package main

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const workers = 50

func main() {
	log.SetLevel(log.TraceLevel)

	ips := make(chan string)

	var wg sync.WaitGroup

	go genIP(ips)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(ips, &wg)
	}

	wg.Wait()
}

func worker(channel chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	client := http.DefaultClient
	client.Timeout = time.Duration(20 * time.Second)

	for {
		ip := <-channel

		if ip == "" {
			log.Info("Exit worker oder so")
			return
		}

		log.WithField("ip", ip).Debug("Trying IP")

		resp, err := http.Get("http://" + ip)

		if err != nil {
			var opError *net.OpError
			if errors.As(err, &opError) {
				log.WithError(opError).Trace("Timeout, I guess?")
			} else {
				log.Error(err)
			}

			continue
		}

		xPoweredBy := resp.Header.Get("X-Powered-By")
		server := resp.Header.Get("Server")

		if strings.Contains(xPoweredBy, "PHP/5") || strings.Contains(xPoweredBy, "php/5") || strings.Contains(server, "PHP/5") || strings.Contains(server, "php/5") {
			log.WithField("ip", ip).Info("Found server with PHP 5")
			// fmt.Println("geil: " + ip)
		}

		resp.Body.Close()
	}
}

func genIP(channel chan string) {

	go func() {
		time.Sleep(time.Duration(rand.Int31n(1000*10)) * time.Millisecond)
		channel <- "103.10.56.21"
		channel <- "128.177.134.252"
	}()

	for i1 := 100; i1 < 255; i1++ {
		for i2 := 0; i2 < 255; i2++ {
			for i3 := 0; i3 < 255; i3++ {
				for i4 := 1; i4 < 255; i4++ {
					ip := strconv.Itoa(i1) + "." + strconv.Itoa(i2) + "." + strconv.Itoa(i3) + "." + strconv.Itoa(i4)
					channel <- ip
				}
			}
		}
	}

	close(channel)
}
