package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	//"net"
	//"net/http"
    "os"
	"strconv"
	//"strings"
	//"time"
)

var verbose = flag.Bool("v", false, "Verbose output for debugging")
var meter = flag.String("meter", "", "Meter reader name and port")
var lastUpload = flag.String("lastupload", "/var/cache/PVUploader/last", "File holding last upload time")

func init() {
	flag.Parse()
}

func main() {
    startTime := previousUpload(*lastUpload)
    saveUpload(*lastUpload, startTime)
}

func saveUpload(last string, t int64) {
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
