package main

import (
	"context"
	"crypto/tls"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/vincent-vinf/quic-shell/pkg/shell"
	"github.com/vincent-vinf/quic-shell/pkg/types"
)

const addr = "localhost:4242"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	conn, err := newQuicConn(ctx)
	if err != nil {
		log.Fatal(err)
	}
	data, err := types.PackMsg(&types.Message{
		ID: "1",
	})
	if err != nil {
		log.Fatal(err)
	}
	err = conn.SendMessage(data)
	if err != nil {
		log.Fatal(err)
	}
	start(ctx, conn)

	cancel()
}

func start(ctx context.Context, conn quic.Connection) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			stream, err := conn.AcceptStream(ctx)
			if err != nil {
				log.Println("conn: ", err)
				continue
			}
			go func(s quic.Stream) {
				log.Println("new shell")
				err := shell.New(s, s)
				if err != nil {
					log.Println(err)
					return
				}
				defer s.Close()
			}(stream)
		}
	}
}

func newQuicConn(ctx context.Context) (quic.Connection, error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{types.AppProtocol},
	}
	return quic.DialAddr(ctx, addr, tlsConf, &quic.Config{
		KeepAlivePeriod: time.Second * 30,
		EnableDatagrams: true,
	})
}

func newStream(ctx context.Context, conn quic.Connection) error {
	//stream, err := conn.AcceptStream(ctx)
	//if err != nil {
	//	return err
	//}
	//defer func() { _ = stream.Close() }()
	//
	//_, err = io.Copy(os.Stdout, stream)
	//if err != nil {
	//	return err
	//}
	time.Sleep(time.Minute)
	err := conn.SendMessage([]byte("abcc"))
	if err != nil {
		return err
	}

	return nil
}
