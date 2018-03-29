package main

import (
    "flag"
    "log"
    "net"
    "os"
    "strings"
)

var gpioPins = flag.String("gpio", "", "GPIO pins to watch")
var fileCounters = flag.String("files", "", "Files to open and read")
var interval = flag.Int("interval", 60, "Interval in seconds")
var maxRecords = flag.Int("maxrecords", 24 * 60, "Maximum number of records")
var unixSocket = flag.String("socket", "/tmp/readmeter",
                             "Name of UNIX domain socket")
var verbose = flag.Bool("v", false, "Verbose output for debugging")
var quiet = flag.Bool("q", false, "Do not log events")


func init() {
    flag.Parse()
}

func main() {
    if len(*gpioPins) != 0 {
        gpioCounters(*gpioPins)
    }
    if len(*fileCounters) != 0 {
        fileCount(*fileCounters)
    }
    os.Remove(*unixSocket)
    addr, err := net.ResolveUnixAddr("unix", *unixSocket)
    if err != nil {
        log.Fatalf("Can't resolve %s: %v", *unixSocket, err)
    }
    l, err := net.ListenUnix("unix", addr)
    if err != nil {
        log.Fatalf("Can't listen on %s: %v", *unixSocket, err)
    }
    ch := make(chan Accum, 50)
    go Counter(ch, *interval)
    Updater(l, ch)
}

// Comma separated filenames to watch.
func fileCount(files string) {
    for _, s := range strings.Split(files, ",") {
        if *verbose {
            log.Printf("Opening %s", s)
        }
        f, err := os.Open(s)
        if err != nil {
            log.Printf("Error opening %s: %v", s, err)
            continue
        }
        count, index := addCounter()
        go func() {
            if *verbose {
                log.Printf("Listening on %s (counter %d) for events",
                            s, index)
            }
            for {
                var b [1]byte
                _, err := f.Read(b[:])
                if (err != nil) {
                    log.Fatalf("Read error on %s: %v", s, err)
                }
                if *verbose {
                    log.Printf("Event on %s for counter %d", s, index)
                }
                count()
            }
        }()
    }
}
