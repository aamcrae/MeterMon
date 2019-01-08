package main

import (
	"flag"
	"fmt"
    "io/ioutil"
	"log"
	"net/http"
    "net/url"
	"strings"
	"time"
)

var netPower = flag.Bool("net", false, "PVoutput Net flag setting")
var apikey = flag.String("apikey", "", "PVoutput API key")
var systemid = flag.String("systemid", "", "PVoutput System ID")
var serverUrl = flag.String("url", "https://pvoutput.org/service/r2/addstatus.jsp", "URL of pvoutput upload")

func upload(start int64, period int64, energy int64) error {
    t := time.Unix(start, 0)
    if *verbose {
        log.Printf("Uploading %v: %d Wh", t, energy)
    }
    val := url.Values{}
    val.Add("d", t.Format("20060102"))
    if *netPower {
        val.Add("n", "1")
    }
    val.Add("t", t.Format("15:04"))
    val.Add("v3", fmt.Sprintf("%d", energy))
    req, err := http.NewRequest("POST", *serverUrl, strings.NewReader(val.Encode()))
    if err != nil {
        log.Printf("NewRequest failed: %v", err)
        return err
    }
    req.Header.Add("X-Pvoutput-Apikey", *apikey)
    req.Header.Add("X-Pvoutput-SystemId", *systemid)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    if *verbose || *dryrun {
        log.Printf("req: %s (size %d)", val.Encode(), req.ContentLength)
        if *dryrun {
            return nil
        }
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Req failed: %v", err)
        return err
    }
    defer resp.Body.Close()
    if *verbose {
        log.Printf("Response is: %s", resp.Status)
    }
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        err := fmt.Errorf("Error: %s: %s", resp.Status, body)
        log.Printf("%v", err)
        return err
    }
    return nil
}
