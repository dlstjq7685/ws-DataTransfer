package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

//https://yalantis.com/blog/how-to-build-websockets-in-go/
const (

	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 100000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	comma   = []byte{','}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {

	webSocketConfig()
}

var addr = flag.String("localhost", ":8080", "http service address")

func main() {
	log.Println("Open WS server")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		ServeWs(w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func webSocketConfig() {

	// websocket SubProtocol Config
	upgrader.Subprotocols = []string{"testticket"}

	// cors solution
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

}

func ServeWs(w http.ResponseWriter, r *http.Request) {

	log.Printf("%v\t%v\t%v\n", r.Host, r.UserAgent(), r.Referer())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		log.Println("print line", time.Now())
		return nil
	})

	log.Println("connect client")

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v¥n", err)
			}
			break
		}

		c.send <- message

		log.Println(message)
	}

	log.Println("end client")

}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {

	// ping frame send ticker
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return
	}

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			fmt.Println("test1 ", string(message))
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}
			fmt.Println("test2 ", message)

			if err := w.Close(); err != nil {

				return
			}

		// ping packet send
		case <-ticker.C:
			log.Println("ping")
			log.Println(time.Now())
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}
