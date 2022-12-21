package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	server := &http.Server{}
	server.Addr = ":8080"
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodConnect:
			handleTunnel(w, r)
		default:
			handleHTTP(w, r)
		}
	})

	log.Println("start server on port", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func handleTunnel(w http.ResponseWriter, r *http.Request) {
	log.Print("new tcp connection", r.Method)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transmit(destConn, clientConn)
	go transmit(clientConn, destConn)
}

func transmit(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	rr, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	if err != nil {
		log.Print(err)
		return
	}
	copyHeader(req.Header, rr.Header)

	var transport http.Transport
	resp, err := transport.RoundTrip(rr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("Resp-Headers: %v\n", resp.Header)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	dH := w.Header()
	copyHeader(resp.Header, dH)
	dH.Add("Requested-Host", rr.Host)

	w.Write(body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
