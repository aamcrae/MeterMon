package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

const (
    Deadline = time.Minute * 5
)

// Client holds context about each client that is connected.
type Client struct {
    send chan []Accum
    done chan struct{}
    vals []Accum
}

// Updater accepts new connections and updates the clients with
// new accumulated counts. The first time that a client connects,
// all available data is sent.
func Updater(l net.Listener, c <-chan Accum) {
    // Circular buffer containing records.
    recs := make ([]Accum, MaxRecords)
    next := 0
    var clients []*Client
    acceptor := make(chan net.Conn)
    go func() {
        for {
            newClient, err := l.Accept()
            if err == nil {
                acceptor <- newClient
            } else {
                fmt.Fprintf(os.Stderr, "Accept err: %v\n", err)
            }
        }
    }()
    dataToSend := false
    for {
        select {
        case newVal := <- c:
            recs[next] = newVal
            next = (next + 1) % MaxRecords
            // Append this new value to the values to be sent to the clients.
            for _, c := range clients {
                c.vals = append(c.vals, newVal)
            }
        case newConn := <- acceptor:
            fmt.Fprintf(os.Stderr, "New connection accepted\n")
            newClient := new(Client)
            newClient.send = make(chan []Accum, 10)
            newClient.done = make(chan struct {}, 1)
            // Copy over the existing data.
            newClient.vals = append(newClient.vals, recs[next:]...)
            if next != 0 {
                newClient.vals = append(newClient.vals, recs[:next]...)
            }
            clients = append(clients, newClient)
            go updateClient(newClient.send, newClient.done, newConn)
        default:
            if (dataToSend) {
                time.Sleep(time.Second)
            } else {
                time.Sleep(time.Second)
            }
        }
        clients, dataToSend = clientUpdates(clients)
    }
}

func clientUpdates(clients []*Client) ([]*Client, bool) {
    dataToSend := false
    // Check for any closed clients.
    for i := 0; i < len(clients); i++ {
        select {
        case <-clients[i].done:
            // Client has timed out or closed, shut down and remove.
            close(clients[i].send)
            // Move last element to replace current element.
            if i != len(clients)-1 {
                clients[i] = clients[len(clients)-1]
            }
            // Nil out the last element to avoid memory leaks,
            // and trim off the last element.
            clients[len(clients)-1] = nil
            clients = clients[:len(clients)-1]
            // Since the current element was removed, decrement the
            // index counter so that the element replacing it is accessed.
            i--
        default:
        }
    }
    // Check for any data to send.
    for _, c := range clients {
        if len(c.vals) != 0 {
            select {
            case c.send <- c.vals:
                // Data sent, so remove the values.
                c.vals = []Accum{}
            default:
                dataToSend = true
            }
        }
    }
    return clients, dataToSend
}

// updateClient reads slices of Accum, formats them, and sends
// the text output to the client as space separated values.
func updateClient(input <-chan []Accum, done chan<- struct{}, conn net.Conn) {
    defer conn.Close()
    defer close(done)
    for {
        for ac := range input {
            for _, val := range ac {
                // Skip uninitialised values.
                if val.interval != 0 {
                    conn.SetWriteDeadline(time.Now().Add(Deadline))
                    if err := writeClient(val, conn); err != nil {
                        fmt.Fprintf(os.Stderr, "Closing client: %v\n", err)
                        return
                    }
                }
            }
        }
    }
}

// writeClient sends one line of output to the client.
func writeClient(ac Accum, conn net.Conn) error {
    if _, err := fmt.Fprintf(conn, "%d %d", ac.start.Unix(), ac.interval); err != nil {
        return err
    }
    for _, v := range ac.counts {
        if _, err := fmt.Fprintf(conn, " %d", v); err != nil {
            return err
        }
    }
    if _, err := fmt.Fprintf(conn, "\n"); err != nil {
        return err
    }
    return nil
}
