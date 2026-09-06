package websocket

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"streamforge/internal/metrics"
)

const RedisPubSubChannel = "ws:events"

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.Mutex

	redisClient *redis.Client
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		redisClient: redisClient,
	}
}

func (h *Hub) Run(ctx context.Context) {
	// Start redis listener for cross-instance broadcasts
	go h.listenRedis(ctx)

	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			metrics.WebsocketConnections.Set(float64(len(h.clients)))
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				metrics.WebsocketConnections.Set(float64(len(h.clients)))
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Bounded channel is full, meaning the client is slow.
					// We disconnect slow clients to prevent blocking the Hub.
					delete(h.clients, client)
					close(client.send)
					metrics.WebsocketConnections.Set(float64(len(h.clients)))
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) listenRedis(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		pubsub := h.redisClient.Subscribe(ctx, RedisPubSubChannel)
		ch := pubsub.Channel()

		for msg := range ch {
			// Forward Redis pubsub messages to local connected clients
			select {
			case h.broadcast <- []byte(msg.Payload):
			case <-ctx.Done():
				pubsub.Close()
				return
			}
		}

		pubsub.Close()

		if ctx.Err() != nil {
			return
		}

		slog.Warn("Redis pubsub subscription lost, reconnecting in 1s...")
		time.Sleep(1 * time.Second)
	}
}
