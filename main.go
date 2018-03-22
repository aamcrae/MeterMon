package main

import (
    "fmt"
    "os"
)

func main() {
    c := make(chan Reading, 50)
    go Counter(c, 5, os.Stdin)
    for {
        r := <- c
        fmt.Printf("Start: %v, intv %v count %v\n", r.start,
                   r.interval, r.count)
    }
}
