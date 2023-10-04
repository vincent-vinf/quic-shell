package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/quic-go/quic-go"
	"golang.org/x/term"
)

const addr = "localhost:4242"

const message = "123"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	err := client(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cancel()
}

func client(ctx context.Context) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(ctx, addr, tlsConf, nil)
	if err != nil {
		return err
	}

	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		_, _ = io.Copy(stream, os.Stdin)
		log.Println("stream -> out: end")
	}()
	_, _ = io.Copy(os.Stdout, stream)
	log.Println("in -> stream: end")

	return nil
}
