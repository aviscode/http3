package main

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go/http3"
	"io"
	"net/http"
)

func main() {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c := http.Client{
		Transport: roundTripper,
	}

	resp, err := c.Get("https://127.0.0.1:4321/")
	if err != nil {
		fmt.Println("cli get err:", err)
	}

	if resp != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("read all err:", err)
		}
		fmt.Println(string(data))
		fmt.Println(resp.Status)
		fmt.Println(resp.Proto)
	}
}
