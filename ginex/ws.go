package ginex

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rickone/athena/common"
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

type WSMessage struct {
	Type int
	Data []byte
}

func JSONWSMessage(val interface{}) *WSMessage {
	data, err := json.Marshal(val)
	common.AssertError(err)

	return &WSMessage{
		Type: websocket.TextMessage,
		Data: data,
	}
}

func WebSocket(f func(ctx context.Context, c *gin.Context, out chan *WSMessage) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer conn.Close()

		out := make(chan *WSMessage)
		defer close(out)

		conn.SetReadLimit(readLimit)
		conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		conn.SetPingHandler(func(data string) error {
			out <- &WSMessage{
				Type: websocket.PongMessage,
			}
			return nil
		})

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case msg := <-out:
					err := conn.WriteMessage(msg.Type, msg.Data)
					if err != nil {
						if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
							GetLogger(c).Errorf("ws.WriteMessage err: %v", err)
						}
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						GetLogger(c).Errorf("ws.ReadMessage err: %v", err)
					}
					break
				}
				conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
			}
			cancel()
		}()

		err = f(ctx, c, out)
		if err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.Canceled {
				GetLogger(c).Errorf("ws.Handler err: %v", err)
			}
		}
	}
}
