package main

import (
    "flag"
    "github.com/davecheney/gpio"
    "github.com/davecheney/gpio/rpi"
    "log"
    "strings"
    "time"
)

const (
  PollTime = 10
)

var debounce = flag.Int("debounce", 100, "Debounce in milliseconds")

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
    deb := time.Duration(*debounce) * time.Millisecond
    pollDuration := time.Duration(PollTime) * time.Millisecond
    lastSample := pin.Get()
    heldState := lastSample
    lastChange := time.Now()
    for {
        time.Sleep(pollDuration)
        p := pin.Get()
        now := time.Now()
        if p != lastSample {
            lastSample = p
            lastChange = now
            continue
        }
        if p != heldState && now.Sub(lastChange) >= deb {
            // Signal has been stable for debounce period.
            // Call counter on rising edge
            if p {
                count()
            }
            if *verbose {
                log.Printf("pin %s now %v\n", s, p)
            }
            heldState = p
        }
    }
}
