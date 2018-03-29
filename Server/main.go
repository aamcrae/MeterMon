package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var unixSocket = flag.String("socket", "/tmp/readmeter",
	"Name of UNIX domain socket")
var verbose = flag.Bool("v", false, "Verbose output for debugging")
var quiet = flag.Bool("q", false, "Do not log events")
var maxrecords = flag.Int("maxrecords", 2880, "Maximum number of records")
var port = flag.Int("port", 80, "Web server port number")

func init() {
	flag.Parse()
}

type Accum struct {
	start    time.Time
	interval int
	counter  []int
}

var counters []Accum
var countersMu sync.Mutex

func main() {
	go ReadData(*unixSocket)
	http.Handle("/", http.HandlerFunc(pageHandler))
	s := &http.Server{Addr: fmt.Sprintf(":%d", *port)}
	log.Fatal(s.ListenAndServe())
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	tparm := r.URL.Query().Get("start")
	var startTime time.Time
	if len(tparm) != 0 {
		if t, err := strconv.ParseInt(tparm, 10, 64); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			startTime = time.Unix(t, 0)
		}
	}
	countersMu.Lock()
	defer countersMu.Unlock()
	for _, ac := range counters {
		if !ac.start.Before(startTime) {
			fmt.Fprintf(w, "%d %d", ac.start.Unix(), ac.interval)
			for _, v := range ac.counter {
				fmt.Fprintf(w, " %d", v)
			}
			w.Write([]byte("\n"))
		}
	}
}

func ReadData(addr string) {
	var latest time.Time
	for {
		conn, err := net.Dial("unix", addr)
		if err != nil {
			if !*quiet {
				log.Printf("Cannot connect to %s: %v", addr, err)
			}
			// Retry after a delay.
			time.Sleep(time.Second * 5)
			continue
		}
		if *verbose {
			log.Printf("Connected to %s", addr)
		}
		r := bufio.NewReader(conn)
		for {
			str, err := r.ReadString('\n')
			if err != nil {
				if !*quiet {
					log.Printf("Read failed on %s: %v", addr, err)
				}
				break
			}
			str = strings.TrimSpace(str)
			// Decode line, which should be in the form:
			//   unix-time interval counter...
			if *verbose {
				log.Printf("Read from %s: \"%s\"", addr, str)
			}
			strs := strings.Split(str, " ")
			if len(strs) < 2 {
				if !*quiet {
					log.Printf("Malformed input on %s: %v", addr, str)
				}
				break
			}
			var a Accum
			if t, err := strconv.ParseInt(strs[0], 10, 64); err != nil {
				if !*quiet {
					log.Printf("Illegal time on %s: %v", addr, str)
				}
				break
			} else {
				a.start = time.Unix(t, 0)
			}
			// Ignore out of date values
			if latest.After(a.start) {
				if *verbose {
					log.Printf("Discarding out-of-date %s", str)
				}
				continue
			}
			latest = a.start
			if intv, err := strconv.ParseUint(strs[1], 10, 0); err != nil {
				if !*quiet {
					log.Printf("Illegal interval on %s: %v", addr, str)
				}
				break
			} else {
				a.interval = int(intv)
			}
			for i, s := range strs[2:] {
				if c, err := strconv.ParseUint(s, 10, 0); err == nil {
					a.counter = append(a.counter, int(c))
				} else if !*quiet {
					log.Printf("Counter %d malformed, ignored, on %s: %v", i, addr, str)
				}
			}
			if *verbose {
				log.Printf("Adding entry: %v %v %v", a.start, a.interval, a.counter)
			}
			addToCounters(a)
		}
	}
}

func addToCounters(a Accum) {
	countersMu.Lock()
	defer countersMu.Unlock()
	if len(counters) >= *maxrecords {
		counters = append(counters[1:], a)
	} else {
		counters = append(counters, a)
	}
}
