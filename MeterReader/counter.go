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
    counts []int64
}

const maxCounters = 32

var counters [maxCounters]int64
var counterIndex int32

// Counter accumulates the event counts during the intervals, and
// sends the accumlated counts down a channel. The counters must all
// be allocated before calling this function.
func Counter(c chan<- Accum, interval int) {
    intv := time.Duration(interval) * time.Second  // Interval as duration.
    numCounters := int(atomic.LoadInt32(&counterIndex))
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
        reading.counts = make([]int64, numCounters)
        for i := 0; i < numCounters; i++ {
            reading.counts[i] = atomic.SwapInt64(&counters[i], 0)
        }
        if *verbose {
            fmt.Println("Start:", reading.start, "Interval:", reading.interval,
                        "counts = ", reading.counts)
        }
        c <- reading
    }
}

// Threadsafe routine to assign a counter.
func addCounter() (func(), int) {
    index := atomic.AddInt32(&counterIndex, 1) - 1
    if index >= maxCounters {
        fmt.Fprintf(os.Stderr, "Maximum number of counters exceeded!\n")
        os.Exit(1)
    }
    counter := &counters[index]
    return func() {
        atomic.AddInt64(counter, 1)
    }, int(index)
}
