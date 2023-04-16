package litws

import (
	"fmt"
	"log"
	"net/http"
)

type LitwsOptions struct {
	Host             string
	Port             uint64
	ReconnectDelayMs uint64
	AuthKey          string
}

type Litws struct {
	Options      *LitwsOptions
	Synchronizer *Synchronizer

	packetHandler map[uint64]packetHandler
	clients       map[*Client]bool
	register      chan *Client
	unregister    chan *Client
}

func NewLitws(options *LitwsOptions) *Litws {
	lws := &Litws{
		Options:       options,
		packetHandler: map[uint64]packetHandler{},
		clients:       make(map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
	lws.Synchronizer = lws.NewSynchronizer()
	lws.Synchronizer.init()
	return lws
}

func (lws *Litws) RegisterPacketHandlers(handlers map[uint64]packetHandler) {
	for id, handler := range handlers {
		lws.packetHandler[id] = handler
	}
}

func (lws *Litws) Send(packet Packet[any]) {
	// todo
}

func (lws *Litws) Run() {

	go func() {
		log.Printf("waiting for connections...")
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			lws.serveWs(w, r)
		})
		if err := http.ListenAndServe(fmt.Sprintf(":%d", lws.Options.Port), nil); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	for {
		select {
		case client := <-lws.register:
			lws.clients[client] = true
		case client := <-lws.unregister:
			if _, ok := lws.clients[client]; ok {
				delete(lws.clients, client)
				close(client.send)
			}
			/*case message := <-lws.broadcast:
			for client := range lws.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(lws.clients, client)
				}
			}*/
		}
	}
}
