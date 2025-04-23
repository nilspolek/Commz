// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"team6-managing.mni.thm.de/Commz/gateway/internal/utils"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

var (
	logger = utils.GetLogger("ws")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	userId  uuid.UUID
	cookie  string
	isAlive bool
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error().Err(err).Str("user-id", c.userId.String()).Msg("unexpected websocket close")
			}
			break
		}
	}
}

// sendError sends an error message back to the client
func (c *Client) sendError(message string) {
	errorMsg := struct {
		Error string `json:"error"`
	}{
		Error: message,
	}

	if bytes, err := json.Marshal(errorMsg); err == nil {
		c.send <- bytes
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.isAlive = false
		logger.Info().Str("user-id", c.userId.String()).Msg("client connection closed")
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				logger.Info().Str("user-id", c.userId.String()).Msg("hub closed client channel")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error().Err(err).Str("user-id", c.userId.String()).Msg("failed to get writer")
				return
			}

			if _, err := w.Write(message); err != nil {
				logger.Error().Err(err).Str("user-id", c.userId.String()).Msg("failed to write message")
				return
			}

			if err := w.Close(); err != nil {
				logger.Error().Err(err).Str("user-id", c.userId.String()).Msg("failed to close writer")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error().Err(err).Str("user-id", c.userId.String()).Msg("ping failed")
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	cookies := r.CookiesNamed("commz-token")
	if len(cookies) == 0 {
		logger.Error().Msg("no auth token provided")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := utils.VerifyToken(cookies[0].Value)
	if err != nil {
		logger.Error().Err(err).Msg("failed to verify token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	client := &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		userId:  user.ID,
		cookie:  cookies[0].String(),
		isAlive: true,
	}

	logger.Info().
		Str("user-id", user.ID.String()).
		Msg("new websocket connection established")

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
