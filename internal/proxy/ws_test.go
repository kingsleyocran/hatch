package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestWSHubBroadcast(t *testing.T) {
	hub := NewWSHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/__hatch/ws", hub.ServeWS)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	addr := ln.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/__hatch/ws", addr), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	hub.Broadcast("ready")

	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(msg) != "ready" {
		t.Errorf("message = %q, want %q", string(msg), "ready")
	}
}

func TestWSHubMultipleClients(t *testing.T) {
	hub := NewWSHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/__hatch/ws", hub.ServeWS)

	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()

	addr := ln.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn1, _, _ := websocket.Dial(ctx, fmt.Sprintf("ws://%s/__hatch/ws", addr), nil)
	conn2, _, _ := websocket.Dial(ctx, fmt.Sprintf("ws://%s/__hatch/ws", addr), nil)
	defer conn1.CloseNow()
	defer conn2.CloseNow()

	hub.Broadcast("ready")

	_, msg1, _ := conn1.Read(ctx)
	_, msg2, _ := conn2.Read(ctx)

	if string(msg1) != "ready" || string(msg2) != "ready" {
		t.Errorf("both clients should receive 'ready'")
	}
}
