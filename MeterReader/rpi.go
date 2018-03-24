package main

import (
    "github.com/davecheney/gpio"
    "github.com/davecheney/gpio/rpi"
    "fmt"
    "os"
    "strings"
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
        if pnum, ok := gpioMap[strings.ToUpper(s)]; ok {
            pin, err := rpi.OpenPin(pnum, gpio.ModeInput)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error opening pin %s! %s\n", s, err)
                defer os.Exit(1)
            }
            count, index := addCounter()
            err = pin.BeginWatch(gpio.EdgeRising, count)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Unable to watch pin %s! %s\n", s, err)
                defer os.Exit(1)
            }
            if *verbose {
                fmt.Printf("Now watching pin %s on counter %d\n", s, index)
            }
        } else {
            fmt.Fprintf(os.Stderr, "Unknown pin: %s\n", s)
            defer os.Exit(1)
        }
    }
}
