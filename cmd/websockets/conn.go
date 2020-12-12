package websockets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
	"strings"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512package websockets

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	Upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

//cleans up the message to send
var (
	newline = "\n"
	space   = " "
)

//The Connection struct represents each Connection to a room
//Ex: If two people access the same room, then the total number of connections would be 2
type Connection struct {
	//websocket Connection
	Ws *websocket.Conn

	//buffered channel for messages
	Send chan *JsonInfo
}

//The Subscription struct is used to register connections to the right room
type Subscription struct {
	Connection *Connection
	Hub        *Hub
	Room       string
}

//Json struct to receive the data from client-side websocket
type JsonInfo struct {
	MessageType   string `json:"message_type"`
	Comment       string `json:"comment"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	PawnPromotion string `json:"pawn_promotion"`
	GameFen       string `json:"game_fen"`
}

//JSON struct to read game data
type GameInfo struct {
	Key       string `json:"Key"`
	Fen       string `json:"Fen"`
	CanChange bool   `json:"CanChange"`
	Expires   string `json:"Expires"`
}

//the message to send and to which room
type Message struct {
	Msg  *JsonInfo
	Room string
}

//this reads messages from the websocket adn pushes them into the hub which then broadcasts them to the other connections in the room
func (s *Subscription) ReadPump() {
	c := s.Connection
	defer func() {
		//our recoverPanic doesn't help if any errors occur and would stop the entire server.
		//In order to alleviate that, we deal with a panic here so that the program can keep running
		if err := recover(); err != nil {
			log.Println(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
		s.Hub.Unregister <- s
		c.Ws.Close()
	}()
	//set readlimit, read deadline time allowed and pong handler
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	//loop that reads messages from the websocket from client slide
	for {
		//make a JsonInfo variable to transfer all the data from the websocket into
		var jsonInfo JsonInfo
		err := c.Ws.ReadJSON(&jsonInfo)
		//check if the websocket message has an error (a regular close counts as an error so we can ignore it)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %e", err)
			}
			break
		}
		//clean up the message if it's a comment
		if jsonInfo.MessageType == "comment" {
			jsonInfo.Comment = strings.TrimSpace(strings.Replace(jsonInfo.Comment, newline, space, -1))
		} else if jsonInfo.MessageType == "send_move" {
			if jsonInfo.GameFen != "" {
				if len(s.Hub.MovesList[s.Room]) > 2 {
					_, s.Hub.MovesList[s.Room] = s.Hub.MovesList[s.Room][0], s.Hub.MovesList[s.Room][1:]
				}
				s.Hub.MovesList[s.Room] = append(s.Hub.MovesList[s.Room], jsonInfo.GameFen)
			}
		}
		//put it into the Message Struct so that the hub knows which room to send the info to
		m := Message{&jsonInfo, s.Room}
		//send it to the Broadcast channel
		s.Hub.Broadcast <- &m
	}
}

func (s *Subscription) WritePump() {
	defer func() {
		//our recoverPanic doesn't help if any errors occur and would stop the entire server.
		//In order to alleviate that, we deal with a panic here so that the program can keep running
		if err := recover(); err != nil {
			log.Println(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()

	//get the connection we want to write to
	c := s.Connection
	//used to keep the connection alive
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	//loop that sends message to the client side websocket
	for {
		c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
		select {
		//get the message from the send channel
		case message, ok := <-c.Send:
			//if err, then close the socket
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			//else we send and check if any error came back or not
			if err := c.Ws.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) write(messageType int, payload []byte) error {
	return c.Ws.WriteMessage(messageType, payload)
}

)

var (
	Upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

//cleans up the message to send
var (
	newline = "\n"
	space   = " "
)

//The Connection struct represents each Connection to a room
//Ex: If two people access the same room, then the total number of connections would be 2
type Connection struct {
	//websocket Connection
	Ws *websocket.Conn

	//buffered channel for messages
	Send chan *JsonInfo

}

//The Subscription struct is used to register connections to the right room
type Subscription struct {
	Connection *Connection
	Hub        *Hub
	Room       string
}

//Json struct to receive the data from client-side websocket
type JsonInfo struct {
	MessageType   string `json:"message_type"`
	Comment       string `json:"comment"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	PawnPromotion string `json:"pawn_promotion"`
	GameFen       string `json:"game_fen"`
}

//JSON struct to read game data
type GameInfo struct {
	Key       string `json:"Key"`
	Fen       string `json:"Fen"`
	CanChange bool   `json:"CanChange"`
	Expires   string `json:"Expires"`
}

//the message to send and to which room
type Message struct {
	Msg  *JsonInfo
	Room string
}

//this reads messages from the websocket adn pushes them into the hub which then broadcasts them to the other connections in the room
func (s *Subscription) ReadPump() {
	c := s.Connection
	defer func() {
		//our recoverPanic doesn't help if any errors occur and would stop the entire server.
		//In order to alleviate that, we deal with a panic here so that the program can keep running
		if err := recover(); err != nil {
			log.Println(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
		s.Hub.Unregister <- s
		c.Ws.Close()
	}()
	//set readlimit, read deadline time allowed and pong handler
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	//loop that reads messages from the websocket from client slide
	for {
		//make a JsonInfo variable to transfer all the data from the websocket into
		var jsonInfo JsonInfo
		err := c.Ws.ReadJSON(&jsonInfo)
		//check if the websocket message has an error (a regular close counts as an error so we can ignore it)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %e", err)
			}
			break
		}
		//clean up the message if it's a comment
		if jsonInfo.MessageType == "comment" {
			jsonInfo.Comment = strings.TrimSpace(strings.Replace(jsonInfo.Comment, newline, space, -1))
		} else if jsonInfo.MessageType == "send_move" {
			if jsonInfo.GameFen != "" {
				if len(s.Hub.MovesList[s.Room]) > 2 {
					_, s.Hub.MovesList[s.Room] = s.Hub.MovesList[s.Room][0], s.Hub.MovesList[s.Room][1:]
					s.Hub.MovesList[s.Room] = append(s.Hub.MovesList[s.Room], jsonInfo.GameFen)
				} else {
					s.Hub.MovesList[s.Room] = append(s.Hub.MovesList[s.Room], jsonInfo.GameFen)
				}
			}
		}
		//put it into the Message Struct so that the hub knows which room to send the info to
		m := Message{&jsonInfo, s.Room}
		//send it to the Broadcast channel
		s.Hub.Broadcast <- &m
	}
}

func (s *Subscription) WritePump() {
	defer func() {
		//our recoverPanic doesn't help if any errors occur and would stop the entire server.
		//In order to alleviate that, we deal with a panic here so that the program can keep running
		if err := recover(); err != nil {
			log.Println(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()

	//get the connection we want to write to
	c := s.Connection
	//used to keep the connection alive
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	//loop that sends message to the client side websocket
	for {
		c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
		select {
		//get the message from the send channel
		case message, ok := <-c.Send:
			//if err, then close the socket
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			//else we send and check if any error came back or not
			if err := c.Ws.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) write(messageType int, payload []byte) error {
	return c.Ws.WriteMessage(messageType, payload)
}
