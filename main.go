package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/go-stomp/stomp/v3/server/client"
	"github.com/go-stomp/stomp/v3/server/topic"
	"github.com/gorilla/websocket"
)

var listenAddr = flag.String("addr", ":8765", "Listen address")
var helpFlag = flag.Bool("help", false, "Show this help text")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// CustomConn is a custom implementation of net.Conn that uses bytes.Buffer
type CustomConn struct {
	readBuffer  chan []byte
	writeBuffer chan []byte
}

func NewCustomConn() *CustomConn {
	return &CustomConn{
		readBuffer:  make(chan []byte, 4096),
		writeBuffer: make(chan []byte, 4096),
	}
}

func (c *CustomConn) Read(b []byte) (n int, err error) {
	data := <-c.readBuffer
	n = copy(b, data)
	if n < len(data) {
		c.readBuffer <- data[n:]
	}
	return n, nil
}

func (c *CustomConn) Write(b []byte) (n int, err error) {
	data := make([]byte, len(b))
	copy(data, b)
	c.writeBuffer <- data
	return len(b), nil
}

func (c *CustomConn) Close() error {
	close(c.readBuffer)
	close(c.writeBuffer)
	return nil
}

func (c *CustomConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *CustomConn) RemoteAddr() net.Addr {
	return &net.IPAddr{}
}

func (c *CustomConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *CustomConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *CustomConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type CConfig struct {
}

func (c *CConfig) Authenticate(login, passcode string) bool {
	return true
}

func (c *CConfig) HeartBeat() time.Duration {
	return time.Second * 5
}

func (c *CConfig) Logger() stomp.Logger {
	return nil
}

var tm = topic.NewManager()

func readFromWebSocket(conn *websocket.Conn, localConn *CustomConn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		localConn.readBuffer <- message
	}
}

func writeToWebSocket(conn *websocket.Conn, localConn *CustomConn) {
	for {
		answer, ok := <-localConn.writeBuffer
		if !ok {
			break
		}
		conn.WriteMessage(websocket.TextMessage, answer)
	}
}

func processClientRequests(ch chan client.Request) {
	for response := range ch {
		switch response.Op {
		case client.SubscribeOp:
			topic := tm.Find(response.Sub.Destination())
			topic.Subscribe(response.Sub)
		case client.UnsubscribeOp:
			topic := tm.Find(response.Sub.Destination())
			topic.Unsubscribe(response.Sub)
		case client.EnqueueOp:
			destination, ok := response.Frame.Header.Contains(frame.Destination)
			if ok {
				topic := tm.Find(destination)
				topic.Enqueue(response.Frame)
			}
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade connection:", err)
		return
	}

	localConn := NewCustomConn()
	ch := make(chan client.Request, 128)
	config := CConfig{}
	client.NewConn(&config, localConn, ch)

	go readFromWebSocket(conn, localConn)
	go writeToWebSocket(conn, localConn)
	processClientRequests(ch)
	localConn.Close()
	conn.Close()
}

func main() {
	flag.Parse()
	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	http.HandleFunc("/ws", handleWebSocket)

	log.Println("listening on", *listenAddr)
	err := http.ListenAndServe(*listenAddr, nil)
	if err != nil {
		log.Fatalf("failed to listen: %s", err.Error())
	}
}
