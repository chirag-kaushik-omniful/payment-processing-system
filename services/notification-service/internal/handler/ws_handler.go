package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/omniful/payment-platform/services/notification-service/internal/hub"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub *hub.Hub
}

func NewWSHandler(h *hub.Hub) *WSHandler {
	return &WSHandler{hub: h}
}

func (h *WSHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/ws", h.HandleWebSocket)
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.Register(conn)
	defer h.hub.Unregister(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
