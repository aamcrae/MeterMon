package main

import (
    "github.com/davecheney/gpio"
    "github.com/davecheney/gpio/rpi"
    "log"
    "strings"
    "time"
)

const (
  PollTime = 20
  Debounce = 100
)

var gpioMap = map[string]int{
    "GPIO17": rpi.GPIO17,
    "GPIO21": rpi.GPIO21,
    "GPIO22": rpi.GPIO22,
    "GPIO23": rpi.GPIO23,
    "GPIO24": rpi.GPIO24,
    "GPIO25": rpi.GPIO25,
    "GPIO27": rpi.GPIO27,
}

func gpioCounters(pins string) {
    for _, s := range strings.Split(pins, ",") {
        go pinWatch(s)
    }
}

func pinWatch(s string) {
    pnum, ok := gpioMap[strings.ToUpper(s)]
    if !ok {
        log.Fatalf("Unknown pin: %s\n", s)
    }
    pin, err := rpi.OpenPin(pnum, gpio.ModeInput)
    if err != nil {
        log.Fatalf("Error opening pin %s! %v", s, err)
    }
    count, index := addCounter()
    if *verbose {
        log.Printf("Now watching pin %s on counter %d\n", s, index)
    }
    deb := time.Duration(Debounce) * time.Millisecond
    pollDuration := time.Duration(PollTime) * time.Millisecond
    last := pin.Get()
    current := last
    lastRead := time.Now()
    for {
        time.Sleep(pollDuration)
        p := pin.Get()
        now := time.Now()
        if p != last {
            last = p
            if now.Sub(lastRead) < deb {
               continue
            }
        }
        if p != current {
            // Signal has been stable for debounce period.
            // Call counter on rising edge
            if p {
                count()
            }
            if *verbose {
                log.Printf("pin %s now %v\n", s, p)
            }
            current = p
        }
    }
}
