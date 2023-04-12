package litws

import (
	"context"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

type LitwsOptions struct {
	Host             string
	Port             uint64
	ReconnectDelayMs uint64
}

type Litws struct {
	Options *LitwsOptions

	packetHandler map[uint64]*packetHandler
	clients       []*Client
	clientsMux    *sync.Mutex
}

func NewLitws(options *LitwsOptions) *Litws {
	return &Litws{
		Options:       options,
		packetHandler: map[uint64]*packetHandler{},
		clients:       []*Client{},
		clientsMux:    &sync.Mutex{},
	}
}

func (lws *Litws) RegisterPacketHandlers(handlers map[uint64]*packetHandler) {
	for id, handler := range handlers {
		lws.packetHandler[id] = handler
	}
}

func (lws *Litws) Serve() {
	l, err := net.Listen("tcp", lws.Options.Host+":"+strconv.FormatUint(lws.Options.Port, 10))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.Printf("Litws listening on %s:%d", lws.Options.Host, lws.Options.Port)

	s := &http.Server{Handler: lws}
	errc := make(chan error, 1)
	go func() {
		errc <- s.Serve(l)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("Litws server error: %v", err)
	case <-sigs:
		log.Printf("Litws server interrupted")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Litws server shutdown error: %v", err)
	}
}

func (lws *Litws) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		log.Printf("Litws accept error: %v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "bruh. idk why this happened")

	client := newClient(lws, c)
	lws.clientsMux.Lock()
	defer lws.clientsMux.Unlock()
	lws.clients = append(lws.clients, client)

	go client.readLoop()
	go client.writeLoop()
}

func (lws *Litws) removeClient(c *Client) {

}
