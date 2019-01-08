package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
    "os"
	"strconv"
)

var verbose = flag.Bool("v", false, "Verbose output for debugging")
var meter = flag.String("meter", "", "Meter reader name and port")
var lastUpload = flag.String("lastupload", "/var/cache/PVUploader/last", "File holding last upload time")
var dryrun = flag.Bool("dryrun", false, "Do not upload, but print upload requests")
var interval = flag.Int("interval", 5, "Interval time in minutes")

func init() {
	flag.Parse()
}

func main() {
    lastUploadTime, accum := previousUpload(*lastUpload)
    req := fmt.Sprintf("http://%s/?start=%d", *meter, lastUploadTime)
    if *verbose {
        log.Printf("Request to %s: %s", *meter, req)
    }
    resp, err := http.Get(req)
    if err != nil {
        log.Printf("Request to %s failed: %v", *meter, err)
        os.Exit(1)
    }
    defer resp.Body.Close()
    scanner := bufio.NewScanner(resp.Body)
    intv := int64(*interval * 60)
    entries := 0
    uploaded := 0
    var currentInterval int64
    for scanner.Scan() {
        var start, count int64
        var interval int
        n, _ := fmt.Sscanf(scanner.Text(), "%d %d %d", &start, &interval, &count)
        if n != 3 {
            log.Printf("Cannot parse: %s", scanner.Text())
            continue
        }
        // Accumulate over the interval time.
        entries++
        if *verbose {
            log.Printf("Entry: %s", scanner.Text())
        }
        if start >= currentInterval + intv {
            if currentInterval != 0 {
                if err := upload(currentInterval, intv, accum); err != nil {
                    break;
                }
                lastUploadTime = currentInterval + intv
                uploaded++
            }
            currentInterval = start - (start % intv)
        }
        accum += count
    }
    if err := scanner.Err(); err != nil {
        log.Printf("Read on %s failed: %v", *meter, err)
        os.Exit(1)
    }
    if *verbose {
        log.Printf("%d entries successfully read, %d uploaded, %d Wh total", entries,
        uploaded, accum)
    }
    if *verbose {
        log.Printf("Last upload time is %d", lastUploadTime)
    }
    if !*dryrun {
        saveUploadTime(*lastUpload, lastUploadTime, accum)
    }
}

func saveUploadTime(last string, t int64, accum int64) {
    f, err := os.Create(last)
    if err != nil {
        log.Printf("%s: Create failed %v", last, err)
        return
    }
    defer f.Close()
    fmt.Fprintf(f, "%d %d\n", t, accum)
}

func previousUpload(last string) (int64, int64) {
    f, err := os.Open(last)
    if err != nil {
        if *verbose {
            log.Printf("%s: %v", last, err)
        }
        return 0, 0
    }
    defer f.Close()
    var val [2]int64
    scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanWords)
    for i, _ := range val {
        if scanner.Scan() {
            if v, err := strconv.ParseInt(scanner.Text(), 10, 64); err != nil {
                if *verbose {
                    log.Printf("%s: Parse error", last)
                }
            } else {
                val[i] = v
            }
        } else {
            if *verbose {
                log.Printf("%s: Nothing in file", last)
            }
            break
        }
    }
    return val[0], val[1]
}
