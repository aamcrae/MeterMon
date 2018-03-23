package main

import (
    "flag"
    "fmt"
    "net"
    "os"
)

var interval = flag.Int("interval", 10, "Interval in seconds")
var maxRecords = flag.Int("maxrecords", 5000, "Maximum number of records")
var unixSocket = flag.String("socket", "/tmp/readmeter",
                             "Name of UNIX domain socket")
var verbose = flag.Bool("v", false, "Verbose output for debugging")
var quiet = flag.Bool("q", false, "Do not log events")


func init() {
    flag.Parse()
}

func main() {
    c := make(chan Accum, 50)
    countFile(os.Stdin)
    countFile(os.Stdin)
    go Counter(c, *interval)
    os.Remove(*unixSocket)
    addr, err := net.ResolveUnixAddr("unix", *unixSocket)
    if err != nil {
        fmt.Printf("Can't resolve %s: %v", *unixSocket, err)
        return
    }
    l, err := net.ListenUnix("unix", addr)
    if err != nil {
        fmt.Printf("Can't listen on %s: %v", *unixSocket, err)
        return
    }
    Updater(l, c)
}
