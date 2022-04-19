package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type myproxy struct {
	protocol string
	ip       string
	port     int
}

func checker(id int, jobs <-chan myproxy,
	liveProxies chan<- string, stat chan<- int,
	wg *sync.WaitGroup, testURL string, proxyGood string, userAgent string) {
	var proxyURL string
	defer wg.Done()
	var tp Transport
	for prox := range jobs {
		proxyURL = fmt.Sprintf("%s://%s:%d", prox.protocol, prox.ip, prox.port)
		tp.init(proxyURL)
		//fmt.Println("tp.init done")
		//fmt.Printf("+ testing %s %s\r", proxyURL, strings.Repeat(" ", len(proxyURL)+5))
		var content string
		var header Header
		var isGood ComboOrProxyIsGood
		var cmb combo
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		header.init(testURL, userAgent)
		respChan := make(chan string, 1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()
			_, _, content = GetRequest(testURL, tp.transport, 13, &header, 5)
			select {
			case <-ctx.Done(): //check if content is no longer needed
				return //prevent lingering goroutine
			case respChan <- content:

			}
		}()
		var contentFromChan string
		hardTimeout := 25
		x := 0
	forL:
		for {
			select {
			case contentFromChan = <-respChan:
				break forL
			case <-time.After(time.Second * 3):
				fmt.Printf("checking: %s      \r", proxyURL)
				x += 3
				if x > hardTimeout {
					fmt.Println("abandoning goroutine with rogue proxy in it....lol")
					cancel()
					close(respChan)
					break forL
				}
			}
		}
		/* if isGood.withHeader(&header).fromCombo(&cmb).fromContent(content).isGood(proxyGood) == false {
			_, _, content = GetRequest(testURL, tp.transport, 20, &header)
		} */
		//fmt.Printf("Yahoo Length: %d, Google Length: %d", len(yahooContent), len(content))
		//fmt.Println(content)
		if isGood.withHeader(&header).fromCombo(&cmb).fromContent(contentFromChan).isGood(proxyGood) {
			fmt.Println(fmt.Sprint(id) + " proxy " + proxyURL + " is valid")
			//fmt.Println(proxyGood, contentFromChan)
			select {
			case liveProxies <- proxyURL:
			case <-time.After(time.Second * 2):
				fmt.Println(fmt.Sprint(id) + " timedout waiting to write proxy to channel")
			}

		} else {
			fmt.Printf("invalid proxy %s responseLength:%d \r", proxyURL, len(contentFromChan))
		}
		stat <- 1
		time.Sleep(time.Microsecond * 100)
	}
}

func statusUpdater(total int, stat <-chan int) {
	x := 1
	for intd := range stat {
		fmt.Printf("checked %d of %d proxies%s\r", x, total, strings.Repeat(" ", 50))
		x += intd
	}
}

//LoadProxies and test them
func LoadProxies(filename string, testURL string, proxyGood string, userAgent string) []string {
	var wg sync.WaitGroup
	proxies := ReadFile(filename)
	//test proxies
	numGreenThreads := 500
	jobs := make(chan myproxy, numGreenThreads*2)
	stats := make(chan int, 200)
	liveProxies := make(chan string, len(proxies))
	//stats printer
	go statusUpdater(len(proxies), stats)
	for w := 1; w <= numGreenThreads; w++ {
		wg.Add(1)
		go checker(w, jobs, liveProxies, stats, &wg, testURL, proxyGood, userAgent)
	}
	//add jobs to queue

	for _, proxyLine := range proxies {
		if !strings.Contains(proxyLine, "SOCKS5") && !strings.Contains(proxyLine, "HTTP") {
			proxyLine = "SOCKS5:" + proxyLine
		}
		parts := strings.Split(proxyLine, ":")
		if len(parts) == 3 {
			job := myproxy{
				protocol: strings.ToLower(parts[0]),
				ip:       parts[1],
				port:     CastToInteger(parts[2])}

			jobs <- job
		}
	}

	close(jobs)
	fmt.Println("all proxies queued")
	wg.Wait()
	fmt.Println("finished checking proxies")
	fmt.Println("Done")
	close(stats)
	close(liveProxies)
	//now read good proxies to list and return it
	validProxies := []string{}
	for prox := range liveProxies {
		validProxies = append(validProxies, prox)
	}
	return validProxies
}
