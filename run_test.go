package main

import (
	"strings"
	"testing"
)

const serverAddress = "http://127.0.0.1:8080"
const token = "longlongauthtoken"

func TestServer(t *testing.T) {
	server(":8080", token)
}

func TestClientTCPByMySQL(t *testing.T) {
	client("0.0.0.0:13308", serverAddress, token, strings.TrimPrefix("127.0.0.1:3308", "http://"), "tcp")
}

func TestClientTCPByRedis(t *testing.T) {
	client("0.0.0.0:16379", serverAddress, token, strings.TrimPrefix("127.0.0.1:6379", "http://"), "tcp")
}

func TestClientUDP(t *testing.T) {
	client("0.0.0.0:19098", serverAddress, token, strings.TrimPrefix("127.0.0.1:9098", "http://"), "udp")
}
