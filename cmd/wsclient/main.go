package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"

	"golang.org/x/term"

	"github.com/vincent-vinf/quic-shell/pkg/util"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	wrapper := util.NewWsWrapper(c)

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal("term:", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		_, err = io.Copy(wrapper, os.Stdin)
	}()
	_, err = io.Copy(os.Stdout, wrapper)
}
