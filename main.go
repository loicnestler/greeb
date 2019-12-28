package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const workers = 1

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

	for {
		ip := <-channel
		log.Debug(ip)

		if ip == "" {
			log.Info("Exit worker oder so")
			return
		}

		resp, err := http.Get("http://" + ip)

		if err != nil {
			log.Fatal(err)
			continue
		}

		log.Print(ip)
		xPoweredBy := resp.Header.Get("X-Powered-By")
		if strings.Contains(xPoweredBy, "PHP/5") || strings.Contains(xPoweredBy, "php/5") {
			fmt.Println(ip)
		}

		resp.Body.Close()
	}
}

func genIP(channel chan string) {
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
