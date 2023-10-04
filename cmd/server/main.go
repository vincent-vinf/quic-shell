package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	err := quicServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cancel()
}

func quicServer(ctx context.Context) error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			return err
		}
		go func() {
			err := handle(ctx, conn)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func handle(ctx context.Context, conn quic.Connection) error {
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()
	err = shell(stream, stream)
	if err != nil {
		return err
	}

	return nil
}

func shell(in io.Reader, out io.Writer) error {
	c := exec.Command("zsh")
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()

	go func() { _, _ = io.Copy(ptmx, in) }()
	_, _ = io.Copy(out, ptmx)

	return nil
}

// Setup a bare-bones TLS config for the client
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
		NextProtos:   []string{"quic-echo-example"},
	}
}
