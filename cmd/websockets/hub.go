package websockets

import (
	"database/sql"
	"fmt"
	"github.com/sheshan1961/chessapp/pkg/models/mysql"
	"log"
	"runtime/debug"
)

type Hub struct {
	//Think of Rooms like {"Room-Key" : {conn_1 : True, conn_2: True}, "Room-Key-2" : {conn_1 : true}}
	Rooms      map[string]map[*Connection]bool
	Broadcast  chan *Message
	Register   chan *Subscription
	Unregister chan *Subscription
	MovesList  map[string][]string
	Game       *mysql.GameModel
}


func NewHub(db *sql.DB) *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Connection]bool),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Subscription),
		Unregister: make(chan *Subscription),
		MovesList:  make(map[string][]string),
		Game:       &mysql.GameModel{DB: db},
	}
}

func (h *Hub) Run() {
	defer func() {
		//our recoverPanic doesn't help if any errors occur and would stop the entire server.
		//In order to alleviate that, we deal with a panic here so that the program can keep running
		if err := recover(); err != nil {
			log.Println(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()
	for {
		select {
		case subscription := <-h.Register:
			//get all current connections in the room that the new subscription is trying to enter
			connections := h.Rooms[subscription.Room]
			//if no connections, make one and add the new connection into the h.Rooms map
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.Rooms[subscription.Room] = connections
			}
			//set the connection bool value in h.Room to true
			h.Rooms[subscription.Room][subscription.Connection] = true
		case subscription := <-h.Unregister:
			//get all current connections in the room
			connections := h.Rooms[subscription.Room]
			//if there are connections in the room, then we need to get rid of them
			if connections != nil {
				//all of the connections should be true but doesn't hurt to check
				if ok := connections[subscription.Connection]; ok {
					//delete the connection from the hub's connection list
					delete(connections, subscription.Connection)
					//close the connection's send channel
					close(subscription.Connection.Send)
					//if there are no more connections, then we delete the room as well
					if len(connections) == 0 {
						h.SaveGame(subscription.Room)
						delete(h.Rooms, subscription.Room)
					}
				}
			}
		case message := <-h.Broadcast:
			//get all the connections currently in the room where we need to send the message
			connections := h.Rooms[message.Room]
			//loop through the connections
			for c := range connections {
				select {
				//if the send channel is open, send a message
				case c.Send <- message.Msg:
				//This happens when the send channel has lots of messages(implying that none are being broadcasted), so we just close it
				default:
					close(c.Send)
					delete(connections, c)
					if len(connections) == 0 {
						delete(h.Rooms, message.Room)
					}
				}
			}
		}
	}
}

func (h *Hub) GetMoves(room string) []string {
	var moveslist []string
	for _, moves := range h.MovesList[room] {
		moveslist = append(moveslist, moves)
	}
	return moveslist
}

func (h *Hub) SaveGame(room string) {
	lastMove := h.MovesList[room][len(h.MovesList[room])-1]
	_ = h.Game.Save(room, lastMove)
}

func (h *Hub) GetLen(room string) int {
	return len(h.Rooms[room])
}