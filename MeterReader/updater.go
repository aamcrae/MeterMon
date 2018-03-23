package main

import (
    "bytes"
    "fmt"
    "net"
    "os"
    "time"
)

const (
    Deadline = time.Minute * 5
    MsgWindow = 5
)

// Client holds context about each client that is connected.
type Client struct {
    send chan []Accum
    vals []Accum
}

type clientMsg struct {
    client *Client
    done bool
}

// Updater accepts new connections and updates the clients with
// new accumulated counts. The first time that a client connects,
// all available data is sent.
func Updater(l net.Listener, c <-chan Accum) {
    // Circular buffer containing records.
    recs := make ([]Accum, *maxRecords)
    next := 0
    clients := make(map[*Client]*Client)
    clientChan := make(chan clientMsg, 100)
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
    for {
        select {
        case newVal := <- c:
            recs[next] = newVal
            next = (next + 1) % *maxRecords
            // Append this new value to the values to be sent to the clients.
            for _, c := range clients {
                c.vals = append(c.vals, newVal)
                sendToClient(c)
            }
        case newConn := <- acceptor:
            if (!*quiet) {
                fmt.Fprintf(os.Stderr, "New connection accepted from %v\n",
                        newConn.RemoteAddr())
            }
            client := new(Client)
            client.send = make(chan []Accum, MsgWindow)
            clients[client] = client
            // Copy over the existing data.
            client.vals = append(client.vals, recs[next:]...)
            if next != 0 {
                client.vals = append(client.vals, recs[:next]...)
            }
            sendToClient(client)
            go updateClient(client, clientChan, newConn)
        case msg := <- clientChan:
            if msg.done {
                if (!*quiet) {
                    fmt.Fprintf(os.Stderr, "Closing connection\n")
                }
                delete(clients, msg.client)
            } else {
                // Client has sent data, more can be sent.
                sendToClient(msg.client)
            }
        }
    }
}

// If there is room to send data, and there is data to be sent,
// send it to the client.
func sendToClient(client *Client) {
    if *verbose {
        fmt.Fprintf(os.Stderr, "Sending to client %d vals, window %d\n",
                    len(client.vals), cap(client.send) - len(client.send))
    }
    if len(client.vals) != 0 && len(client.send) < cap(client.send) {
        client.send <- client.vals
        client.vals = []Accum{}
    }
}

// updateClient reads slices of Accum, formats them, and sends
// the text output to the client as space separated values.
func updateClient(client *Client, msgChan chan<- clientMsg, conn net.Conn) {
    defer conn.Close()
    for {
        for ac := range client.send {
            for _, val := range ac {
                // Skip uninitialised values.
                if val.interval != 0 {
                    conn.SetWriteDeadline(time.Now().Add(Deadline))
                    if err := writeClient(val, conn); err != nil {
                        if (!*quiet) {
                            fmt.Fprintf(os.Stderr, "Closing client: %v\n", err)
                        }
                        msgChan <- clientMsg{client, true}
                        return
                    }
                }
            }
            msgChan <- clientMsg{client, false}
        }
    }
}

// writeClient sends one line of output to the client.
func writeClient(ac Accum, conn net.Conn) error {
    var buf bytes.Buffer
    fmt.Fprintf(&buf, "%d %d", ac.start.Unix(), ac.interval)
    for _, v := range ac.counts {
        fmt.Fprintf(&buf, " %d", v)
    }
    buf.WriteString("\n")
    if *verbose {
        fmt.Fprintf(os.Stderr, "Sending to %v: %s", conn.RemoteAddr(), buf.String())
    }
    _, err := conn.Write(buf.Bytes())
    return err
}
