package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	authHeader        = `X-Http2tcp-Auth`
	httpHeaderUpgrade = `http2tcp/1.0`
)

func server(listen string, token string) {
	auth := token[:3]
	key := GenerateKey(token)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") != httpHeaderUpgrade {
			log.Println(r.RemoteAddr, `upgrade failed`)
			http.Error(w, `Upgrade failed`, http.StatusBadRequest)
			return
		}

		if r.Header.Get(authHeader) != auth {
			log.Println(r.RemoteAddr, `auth failed`)
			http.Error(w, `Auth failed`, http.StatusUnauthorized)
			return
		}

		mode := r.Header.Get("mode")
		fmt.Println("mode:" + mode)

		if len(mode) == 0 {
			log.Println(r.RemoteAddr, `mode failed`)
			http.Error(w, `mode failed`, http.StatusBadRequest)
			return
		}

		target, err := DecryptFromBase64(r.URL.Query().Get("target"), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if strings.EqualFold("udp", mode) {
			handleUdp(target, w, r)
			return
		}

		remote, err := net.Dial(`tcp`, target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		remoteCloser := &OnceCloser{Closer: remote}
		defer remoteCloser.Close()

		w.Header().Add(`Content-Length`, `0`)
		w.WriteHeader(http.StatusSwitchingProtocols)

		local, bio, err := w.(http.Hijacker).Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		localCloser := &OnceCloser{Closer: local}
		defer localCloser.Close()

		log.Println(r.RemoteAddr, `->`, target, `connected`)
		defer log.Println(r.RemoteAddr, `->`, target, `closed`)

		if err := bio.Writer.Flush(); err != nil {
			return
		}

		bridge(remote, remoteCloser, local, bio.Reader, localCloser)
	})
	http.ListenAndServe(listen, nil)
}

func handleUdp(target string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleUdp")
	if r.Body == nil {
		fmt.Println(" request body is null")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	fmt.Println(data)
	if err != nil {
		fmt.Println("read request body is error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	udpAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	socket, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	defer socket.Close()
	_, err = socket.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
}
