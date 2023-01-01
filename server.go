package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alta/insecure"
	"github.com/lucas-clemente/quic-go/http3"
)

func insecureLocalCert(addr string) (tls.Certificate, error) {
	sans := insecure.LocalSANs()
	san, _ := os.Hostname()
	if san != "" {
		san = strings.ToLower(san)
		sans = append(sans, san)
	}

	san, _, _ = net.SplitHostPort(addr)
	if san != "" {
		san = strings.ToLower(san)
		sans = append(sans, san)
	}
	return insecure.Cert(sans...)
}

func main() {
	addr := "localhost:4433"
	// Open the listeners
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf(err.Error())
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer udpConn.Close()

	cert, err := insecureLocalCert(addr)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h3", "h3-29"},
		// MinVersion:   tls.VersionTLS13,
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hi im http3 server"))
	}))
	server := http3.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
		Handler:   mux,
	}
	qErr := make(chan error)
	sigChan := make(chan os.Signal, 1)

	go func() {
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		qErr <- server.Serve(udpConn)
	}()
	log.Printf("Server listening at: https://%s", addr)

	select {
	case err := <-qErr:
		// Cannot close the HTTP server or wait for requests to complete properly :/
		log.Fatalf(err.Error())
	case <-sigChan:
		if err := server.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}
	}
}
