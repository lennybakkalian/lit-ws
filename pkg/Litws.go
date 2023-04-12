package litws

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

type LitwsOptions struct {
	Host             string
	Port             uint64
	ReconnectDelayMs uint64
}

type Litws struct {
	Options *LitwsOptions
}

func NewLitws(options *LitwsOptions) *Litws {
	return &Litws{options}
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
	log.Printf("Litws request received: %s", r.URL.Path)
}
