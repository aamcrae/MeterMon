package main

import (
    "fmt"
    "os"
)

func main() {
    c := make(chan Accum, 50)
    go Counter(c, 10, os.Stdin)
    for {
        r := <- c
        fmt.Printf("Start: %v, intv %v count %v\n", r.start.Unix(),
                   r.interval, r.count)
    }
}
