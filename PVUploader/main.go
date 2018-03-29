package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	//"net"
	"net/http"
    "os"
	"strconv"
	//"strings"
	//"time"
)

var verbose = flag.Bool("v", false, "Verbose output for debugging")
var meter = flag.String("meter", "", "Meter reader name and port")
var lastUpload = flag.String("lastupload", "/var/cache/PVUploader/last", "File holding last upload time")
var dryrun = flag.Bool("dryrun", false, "Do not upload, but print upload requests")
var interval = flag.String("interval", 15, "Interval time in minutes")

func init() {
	flag.Parse()
}

func main() {
    startTime := previousUpload(*lastUpload)
    req := fmt.Sprintf("http://%s/?start=%d", *meter, startTime)
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
    entries := 0
    for scanner.Scan() {
        var start int64
        var interval, count int
        n, _ := fmt.Sscanf(scanner.Text(), "%d %d %d", &start, &interval, &count)
        if n != 3 {
            log.Printf("Cannot parse: %s", scanner.Text())
            continue
        }
        entries++
        upload(start, interval, count)
    }
    if err := scanner.Err(); err != nil {
        log.Printf("Read on %s failed: %v", *meter, err)
        os.Exit(1)
    }
    if *verbose {
        log.Printf("%d entries successfully read", entries)
    }
    if !*dryrun {
        saveUploadTime(*lastUpload, startTime)
    }
}

func upload(start int64, interval int, count int) {
}

func saveUploadTime(last string, t int64) {
    f, err := os.Create(last)
    if err != nil {
        log.Printf("%s: Create failed %v", last, err)
        return
    }
    defer f.Close()
    fmt.Fprintf(f, "%d\n", t)
}

func previousUpload(last string) int64 {
    f, err := os.Open(last)
    if err != nil {
        if *verbose {
            log.Printf("%s: %v", last, err)
        }
        return 0
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanWords)
    if scanner.Scan() {
        if start, err := strconv.ParseInt(scanner.Text(), 10, 64); err != nil {
            if *verbose {
                log.Printf("%s: Parse error", last)
            }
        } else {
            return start
        }
    } else {
        if *verbose {
            log.Printf("%s: Nothing in file", last)
        }
    }
    return 0
}
