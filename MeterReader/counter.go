package main

import (
    "fmt"
    "os"
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

// Counter accumulates the event counts during the intervals, and
// sends the accumlated counts down a channel.
func Counter(c <-chan Accum, interval int, files []*os.File) {
    counters := make([]int32, len(files))
    for i, f := range files {
        go countPulses(f, &counters[i])
    }
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
        for i, _ := range counters {
            reading.counts = append(reading.counts, atomic.SwapInt32(&counters[i], 0))
        }
        c <- reading
    }
}

// countPulses reads a file, and increments a counter for each
// byte that is read.
func countPulses(fd *os.File, counter *int32) {
    for {
        var b [1]byte
        _, err := fd.Read(b[:])
        if (err != nil) {
            fmt.Fprintf(os.Stderr, "Error reading from file, %v\n", err)
            os.Exit(1)
        }
        atomic.AddInt32(counter, 1)
    }
}
