package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type WSHub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *WSHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.CloseNow()
	}()

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			return
		}
	}
}

func (h *WSHub) Broadcast(msg string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn.Write(ctx, websocket.MessageText, []byte(msg))
		cancel()
	}
}
