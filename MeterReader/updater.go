package main

import (
    "fmt"
    "net"
    "os"
)

// Updater accepts new connections and updates the clients with
// new accumulated counts. The first time that a client connects,
// all available data is send.
func Updater(l Listener, c chan Accum) {
    c := make(chan Accum, 50)
    for {
        r := <- c
        fmt.Printf("Start: %v, intv %v:", r.start, r.interval)
        for _, v := range r.counts {
            fmt.Printf(" %v", v)
        }
        fmt.Printf("\n")
    }
}
