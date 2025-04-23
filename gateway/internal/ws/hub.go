// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"github.com/google/uuid"
)

type BroadCastMessage struct {
	Bytes    []byte
	Receiver []uuid.UUID
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan BroadCastMessage

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Client count metrics
	metrics struct {
		activeConnections int
		messagesSent      int64
	}
}

func New() *Hub {
	return &Hub{
		broadcast:  make(chan BroadCastMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.metrics.activeConnections++
			logger.Info().Int("active_connections", h.metrics.activeConnections).Msg("client registered")

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.metrics.activeConnections--
				logger.Info().Int("active_connections", h.metrics.activeConnections).Msg("client unregistered")
			}

		case message := <-h.broadcast:
			// Create a map of target users for O(1) lookup
			receivers := make(map[string]bool)
			for _, id := range message.Receiver {
				receivers[id.String()] = true
			}

			// Broadcast message only to intended recipients
			for client := range h.clients {
				if receivers[client.userId.String()] && client.isAlive {
					select {
					case client.send <- message.Bytes:
						h.metrics.messagesSent++
						logger.Debug().
							Str("user-id", client.userId.String()).
							Int64("total_messages", h.metrics.messagesSent).
							Msg("message sent")
					default:
						logger.Warn().
							Str("user-id", client.userId.String()).
							Msg("client send buffer full, closing connection")
						close(client.send)
						delete(h.clients, client)
						h.metrics.activeConnections--
					}
				}
			}
		}
	}
}

// GetMetrics returns current hub metrics
func (h *Hub) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"active_connections": h.metrics.activeConnections,
		"messages_sent":      h.metrics.messagesSent,
	}
}
