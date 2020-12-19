package ginex

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	readLimit           = 64 * 1024
	healthCheckInterval = 20 * time.Second
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func WebSocket(f func(ctx context.Context, c *gin.Context, conn *websocket.Conn) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		conn.SetReadLimit(readLimit)
		conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		conn.SetPingHandler(func(data string) error {
			err := conn.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				return err
			}
			return conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := f(ctx, c, conn); err != nil {
				if s, ok := status.FromError(err); !ok || s.Code() != codes.Canceled {
					GetLogger(c).Error(err)
				}
			}
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					GetLogger(c).Errorf("ReadMessage err: %v", err)
				}
				break
			}
			conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		}
	}
}
