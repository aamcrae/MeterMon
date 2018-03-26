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

// clientMsg is the structure that gets sent back to the producer
// when the client worker has sent a message (or detects that the
// client connection has closed).
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
    clients := make(map[*Client]struct{})
    clientChan := make(chan clientMsg, 100)
    acceptor := make(chan net.Conn)
    go func() {
        for {
            newClient, err := l.Accept()
            if err == nil {
                acceptor <- newClient
            } else if ! *quiet {
                fmt.Fprintf(os.Stderr, "Accept err: %v\n", err)
            }
        }
    }()
    // Main select loop.
    for {
        select {
        case newVal := <- c:
            // New counter values.
            // Append to to the values to be sent to the clients.
            recs[next] = newVal
            next = (next + 1) % *maxRecords
            for c, _ := range clients {
                c.vals = append(c.vals, newVal)
                sendToClient(c)
            }
        case newConn := <- acceptor:
            // New client connection.
            if (!*quiet) {
                fmt.Fprintf(os.Stderr, "New connection accepted from %v\n",
                        newConn.RemoteAddr())
            }
            client := new(Client)
            client.send = make(chan []Accum, MsgWindow)
            clients[client] = struct{}{}
            // Initially send the existing data.
            client.vals = append(client.vals, recs[next:]...)
            if next != 0 {
                client.vals = append(client.vals, recs[:next]...)
            }
            sendToClient(client)
            go clientWorker(client, clientChan, newConn)
        case msg := <- clientChan:
            // Message from the client workers.
            if msg.done {
                // Client has disconnected.
                if (!*quiet) {
                    fmt.Fprintf(os.Stderr, "Closing connection\n")
                }
                close(msg.client.send)
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

// clientWorker reads slices of Accum, formats them, and sends
// the text output to the client as space separated values.
func clientWorker(client *Client, msgChan chan<- clientMsg, conn net.Conn) {
    defer conn.Close()
    for {
        ac := <- client.send
        for _, val := range ac {
            // Skip uninitialised values.
            if val.interval != 0 {
                var buf bytes.Buffer
                fmt.Fprintf(&buf, "%d %d", val.start.Unix(), val.interval)
                for _, v := range val.counts {
                    fmt.Fprintf(&buf, " %d", v)
                }
                buf.WriteString("\n")
                if *verbose {
                    fmt.Fprintf(os.Stderr, "Sending to %v: %s", conn.RemoteAddr(), buf.String())
                }
                conn.SetWriteDeadline(time.Now().Add(Deadline))
                if _, err := conn.Write(buf.Bytes()); err != nil {
                    if (!*quiet) {
                        fmt.Fprintf(os.Stderr, "Closing client: %v\n", err)
                    }
                    msgChan <- clientMsg{client, true}
                    return
                }
            }
        }
        // Send a message indicating the data has been successfully sent.
        msgChan <- clientMsg{client, false}
    }
}
