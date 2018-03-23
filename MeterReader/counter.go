package main

import (
    "fmt"
    "sync/atomic"
    "time"
)

// Accum contains the number of events counted during the interval,
// as well as the start time of the interval.
type Accum struct {
    start time.Time
    interval int
    counts []int32
}

var counters []int32

// Counter accumulates the event counts during the intervals, and
// sends the accumlated counts down a channel.
func Counter(c chan<- Accum, interval int) {
    intv := time.Duration(interval) * time.Second  // Interval as duration.
    for {
        var reading Accum
        // Save current time, rounded to a second.
        now := time.Now()
        reading.start = now.Round(time.Second)
        // Calculate the duration to the end of the interval.
        togo := now.Add(intv).Truncate(intv).Sub(now)
        // Ensure that at least a second will elapsed.
        if (togo < time.Second/2) {
            time.Sleep(time.Second/2)
            continue
        }
        time.Sleep(togo)
        reading.interval = int(time.Now().Sub(now).Round(time.Second).Seconds())
        reading.counts = make([]int32, len(counters))
        for i, _ := range counters {
            reading.counts[i] = atomic.SwapInt32(&counters[i], 0)
        }
        if *verbose {
            fmt.Println("Start:", reading.start, "Interval:", reading.interval,
                        "counts = ", reading.counts)
        }
        c <- reading
    }
}

func addCounter() func() {
    counters = append(counters, 0)
    counter := &counters[len(counters)-1]
    return func() {
        atomic.AddInt32(counter, 1)
    }
}
