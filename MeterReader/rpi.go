package main

import (
    "github.com/davecheney/gpio"
    "fmt"
    "os"
    "strings"
)

var gpioMap = map[string]int{
    "GPIO0": gpio.GPIO0,
    "GPIO1": gpio.GPIO1,
    "GPIO2": gpio.GPIO2,
    "GPIO3": gpio.GPIO3,
    "GPIO4": gpio.GPIO4,
    "GPIO7": gpio.GPIO7,
    "GPIO8": gpio.GPIO8,
    "GPIO9": gpio.GPIO9,
    "GPIO10": gpio.GPIO10,
    "GPIO11": gpio.GPIO11,
    "GPIO17": gpio.GPIO17,
    "GPIO18": gpio.GPIO18,
    "GPIO22": gpio.GPIO22,
    "GPIO23": gpio.GPIO23,
    "GPIO24": gpio.GPIO24,
    "GPIO25": gpio.GPIO25,
}

func gpioCounters(pins string) {
    for _, s := range strings.Split(pins, ",") {
        if pnum, ok := gpioMap[strings.ToUpper(s)]; ok {
            pin, err := gpio.OpenPin(pnum, gpio.ModeInput)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error opening pin %s! %s\n", s, err)
                defer os.Exit(1)
            }
            err = pin.BeginWatch(gpio.EdgeRising, addCounter())
            if err != nil {
                fmt.Fprintf(os.Stderr, "Unable to watch pin %s! %s\n", s, err)
                defer os.Exit(1)
            }
            if *verbose {
                fmt.Printf("Now watching pin %s\n", s)
            }
        } else {
            fmt.Fprintf(os.Stderr, "Unknown pin: %s\n", s)
            defer os.Exit(1)
        }
    }
}
