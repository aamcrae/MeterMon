package main

import (
    "flag"
    "fmt"
    "net"
    "os"
    "strings"
)

var gpioPins = flag.String("gpio", "", "GPIO pins to watch")
var fileCounters = flag.String("files", "", "Files to open and read")
var interval = flag.Int("interval", 300, "Interval in seconds")
var maxRecords = flag.Int("maxrecords", 5000, "Maximum number of records")
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
        fmt.Printf("Can't resolve %s: %v", *unixSocket, err)
        os.Exit(1)
    }
    l, err := net.ListenUnix("unix", addr)
    if err != nil {
        fmt.Printf("Can't listen on %s: %v", *unixSocket, err)
        os.Exit(1)
    }
    ch := make(chan Accum, 50)
    go Counter(ch, *interval)
    Updater(l, ch)
}

// Comma separated filenames to watch.
func fileCount(files string) {
    for _, s := range strings.Split(files, ",") {
        if *verbose {
            fmt.Fprintf(os.Stderr, "Opening %s\n", s)
        }
        f, err := os.Open(s)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", s, err)
            continue
        }
        count, index := addCounter()
        go func() {
            if *verbose {
                fmt.Printf("Listening on %s (counter %d) for events\n",
                            s, index)
            }
            for {
                var b [1]byte
                _, err := f.Read(b[:])
                if (err != nil) {
                    fmt.Fprintf(os.Stderr, "Read error on %s: %v\n", s, err)
                    os.Exit(1)
                }
                if *verbose {
                    fmt.Printf("Event on %s for counter %d\n", s, index)
                }
                count()
            }
        }()
    }
}
