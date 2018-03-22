package main

import (
    "fmt"
    "net"
    "os"
)

constant (
    Interval = 10
    Records = 24 * 60 * 60 / Interval // Keep at least 24 hours of records.
    UnixAddr = "/tmp/readmeter"
)

func main() {
    c := make(chan Accum, 50)
    go Counter(c, Interval, []*os.File{os.Stdin, os.Stdin})
    addr, err := net.ResolveUnixAddr("unix", UnixAddr)
    if err != nil {
        fmt.Printf("Can't resolve %s: %v", UnixAddr, err)
        return
    }
    l, err := net.UnixListener("unix", addr)
    if err != nil {
        fmt.Printf("Can't listen on %s: %v", UnixAddr, err)
        return
    }
    Updater(l, c)
}
