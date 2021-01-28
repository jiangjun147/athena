package ginex

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

type WebSocket struct {
	conn   *websocket.Conn
	read   chan *WSMessage
	write  chan *WSMessage
	ctx    context.Context
	cancel context.CancelFunc
}

type WSMessage struct {
	Type int
	Data []byte
}

func NewWebSocket(c *gin.Context) (*WebSocket, error) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	read := make(chan *WSMessage)
	write := make(chan *WSMessage)

	conn.SetReadLimit(readLimit)
	conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
	conn.SetPingHandler(func(data string) error {
		write <- &WSMessage{
			Type: websocket.PongMessage,
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	ws := &WebSocket{
		conn:   conn,
		read:   read,
		write:  write,
		ctx:    ctx,
		cancel: cancel,
	}

	go ws.doRead()
	go ws.doWrite()
	return ws, nil
}

func (ws *WebSocket) doRead() {
	log.Println("doRead begin")
	defer log.Println("doRead end")

	for {
		msgType, data, err := ws.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("ws.ReadMessage err: %v", err)
			}
			break
		}
		ws.conn.SetReadDeadline(time.Now().Add(healthCheckInterval))

		if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
			ws.read <- &WSMessage{
				Type: msgType,
				Data: data,
			}
		}
	}
	ws.Cancel()
}

func (ws *WebSocket) doWrite() {
	log.Println("doWrite begin")
	defer log.Println("doWrite end")

	for {
		select {
		case msg := <-ws.write:
			err := ws.conn.WriteMessage(msg.Type, msg.Data)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logrus.Errorf("ws.WriteMessage err: %v", err)
				}
				return
			}
		case <-ws.ctx.Done():
			return
		}
	}
}

func (ws *WebSocket) Read() <-chan *WSMessage {
	return ws.read
}

func (ws *WebSocket) Write(msgType int, data []byte) {
	ws.write <- &WSMessage{
		Type: msgType,
		Data: data,
	}
}

func (ws *WebSocket) WriteJSON(val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	ws.Write(websocket.TextMessage, data)
	return nil
}

func (ws *WebSocket) Cancel() {
	ws.cancel()
}

func (ws *WebSocket) Close() {
	ws.conn.Close()
}

func (ws *WebSocket) Context() context.Context {
	return ws.ctx
}

func WSWrap(f func(c *gin.Context, ws *WebSocket) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := NewWebSocket(c)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer ws.Close()

		err = f(c, ws)
		if err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.Canceled {
				GetLogger(c).Errorf("WSWrap err: %v", err)
			}
		}
		ws.Cancel()
	}
}
