package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"io"
	"log"
	"math/big"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/quic-go/quic-go"

	"github.com/vincent-vinf/quic-shell/pkg/manager"
	"github.com/vincent-vinf/quic-shell/pkg/types"
	"github.com/vincent-vinf/quic-shell/pkg/util"
)

const addr = "localhost:4242"

var httpAddr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

var mgr = manager.New()

func main() {
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := quicServer(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.HandleFunc("/echo", echo)
	ListenAndServe(ctx, http.DefaultServeMux, *httpAddr)

	cancel()
}

func quicServer(ctx context.Context) error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), &quic.Config{
		EnableDatagrams: true,
	})
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return err
		}
		go func() {
			err := mgr.Handle(ctx, conn)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{types.AppProtocol},
	}
}

//func () error {
//	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
//	if err != nil {
//		return err
//	}
//	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
//
//	go func() {
//		_, _ = io.Copy(stream, os.Stdin)
//		log.Println("stream -> out: end")
//	}()
//	_, _ = io.Copy(os.Stdout, stream)
//	log.Println("in -> stream: end")
//}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	wrapper := util.NewWsWrapper(c)

	client := mgr.GetClient("1")
	if c != nil {
		stream, err := client.NewStream()
		if err != nil {
			log.Println("open stream", err)
		}
		go func() {
			_, _ = io.Copy(stream, wrapper)
		}()
		_, _ = io.Copy(wrapper, stream)
	} else {
		log.Println("client offline")
	}
}

func ListenAndServe(ctx context.Context, h http.Handler, addr string) {
	srv := &http.Server{
		Addr:    addr,
		Handler: h,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	<-ctx.Done()
}
