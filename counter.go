package main

import (
    "fmt"
    "os"
    "sync/atomic"
    "time"
)

type Reading struct {
    start time.Time
    interval int
    count int32
}

func Counter(c chan Reading, interval int, fd *os.File) {
    var counter int32
    var reading Reading
    intv := time.Duration(interval) * time.Second  // Interval as duration.
    go countPulses(fd, &counter)
    for {
        // Save current time, rounded to a second.
        now := time.Now()
        reading.start = now.Round(time.Second)
        // Calculate the duration to the end of the interval.
        togo := now.Add(intv).Truncate(intv).Sub(now)
        time.Sleep(togo)
        reading.count = atomic.SwapInt32(&counter, 0)
        reading.interval = int(time.Now().Sub(now).Round(time.Second).Seconds())
    }
}

func countPulses(fd *os.File, counter *int32) {
    for {
        var b [1]byte
        _, err := fd.Read(b[:])
        if (err != nil) {
            fmt.Printf("Error reading from file, %v", err)
            return
        }
        atomic.AddInt32(counter, 1)
    }
}
