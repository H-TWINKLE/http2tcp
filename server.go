package main

import (
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
		// 远程ip地址
		remoteAddr := r.Header.Get("RemoteAddr")

		if r.Header.Get("Upgrade") != httpHeaderUpgrade {
			log.Println(r.RemoteAddr, `upgrade failed`, "remoteAddr", remoteAddr)
			http.Error(w, `Upgrade failed`, http.StatusBadRequest)
			return
		}

		if r.Header.Get(authHeader) != auth {
			log.Println(r.RemoteAddr, `auth failed`, "remoteAddr", remoteAddr)
			http.Error(w, `Auth failed`, http.StatusUnauthorized)
			return
		}

		mode := r.Header.Get("mode")
		// fmt.Println("mode:" + mode)

		if len(mode) == 0 {
			log.Println(r.RemoteAddr, `mode failed`, "remoteAddr", remoteAddr)
			http.Error(w, `mode failed`, http.StatusBadRequest)
			return
		}

		target, err := DecryptFromBase64(r.URL.Query().Get("target"), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// udp
		if strings.EqualFold("udp", mode) {
			handleUdp(target, mode, remoteAddr, w, r)
			return
		}

		//tcp
		handleTcp(w, r, err, target, mode, remoteAddr)
	})
	http.ListenAndServe(listen, nil)
}

func handleTcp(w http.ResponseWriter, r *http.Request, err error, target string, mode string, remoteAddr string) {
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

	log.Println(r.RemoteAddr, `->`, target, `connected`, `mode`, mode)
	defer log.Println(r.RemoteAddr, `->`, target, `closed`, `mode`, mode)

	if err := bio.Writer.Flush(); err != nil {
		return
	}

	bridge(remote, remoteCloser, local, bio.Reader, localCloser)
}

func handleUdp(target string, mode string, remoteAddr string, w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		log.Println(" request body is null", "remoteAddr", remoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	// log.Println("read udp mode", "remoteAddr", remoteAddr, " request body is ", string(data))

	if err != nil {
		log.Println("read request body is error", "remoteAddr", remoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 获取addr地址
	udpAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		log.Println("remoteAddr", remoteAddr, "error ResolveUDPAddr udp:", err)
		return
	}
	socket, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("remoteAddr", remoteAddr, "error DialUDP udp:", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	defer socket.Close()

	log.Println(r.RemoteAddr, `->`, target, `connected`, `mode`, mode)
	defer log.Println(r.RemoteAddr, `->`, target, `closed`, `mode`, mode)

	_, err = socket.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	// 和客户端统一
	w.WriteHeader(http.StatusSwitchingProtocols)
}
