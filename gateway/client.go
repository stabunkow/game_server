package main

import (
	"context"
	"game_server/model"
	"game_server/pb"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Each client as conn, with ctx to control gorotine cancel
type Client struct {
	ctx       context.Context
	cancel    func()
	closeOnce sync.Once
	conn      *websocket.Conn
	hub       *Hub
	send      chan []byte
	sid       string
	uid       string
}

func newClient(conn *websocket.Conn, hub *Hub, uid string) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte),
		uid:    uid,
	}

	client.hub.register <- client

	go client.readPump()
	go client.writePump()

	return client
}

// Use proto3 to unserialize data, and forward to service
func (c *Client) readPump() {
	defer func() {
		c.Exit()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	c.conn.SetCloseHandler(func(int, string) error { c.Exit(); return nil })

	for {
		select {
		case <-c.ctx.Done(): // exit gorotine
			return
		default: // handle msg
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println(err)
				}
				return
			}

			// proto unserialize
			msg := &pb.Message{}
			err = proto.Unmarshal(data, msg)
			if err != nil {
				log.Println(err)
				continue
			}

			// forward to service, only accept Req and Notify, as entrypoint
			// use gorotine to handle sync as async ## !important
			switch msg.GetMessage().(type) {
			case *pb.Message_Req:
				// each rsp use same mid from req
				switch req := msg.GetReq(); req.GetReq().(type) {
				case *pb.Req_GetUserInfoReq:
					go c.GetUserInfo(req)
				}
			case *pb.Message_Notify:
				switch ntf := msg.GetNotify(); ntf.GetNotify().(type) {
				case *pb.Notify_ChatNotify:
					go c.Chat(ntf)
				}
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Exit()
	}()

	for {
		select {
		case <-c.ctx.Done(): // exit gorotine
			return
		case <-ticker.C: // handle ping
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case message, ok := <-c.send: // send
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) Exit() {
	c.closeOnce.Do(func() {
		// Save redis data to mgo, and clear
		if model.UserStorageExists(c.uid) {
			usr := model.LoadUserById(c.uid)
			usr.Save()
			usr.ClearStorage()
		}

		c.hub.unregister <- c
		c.conn.Close()
		c.cancel()
	})
}

// Send close control with reason
func (c *Client) ExitWithReason(reason string) {
	defer c.Exit()
	err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason), time.Now().Add(time.Second))
	if err != nil {
		return
	}
}

// handle req
func (c *Client) GetUserInfo(req *pb.Req) {
	arg := &pb.String{Value: c.uid}
	reply, err := GetGameServiceClient().GetUserInfo(context.TODO(), arg)
	rsp := &pb.Message{}
	if err != nil {
		rsp = pb.MakeRsp_Error(req.GetMid(), err.Error())
	} else {
		rsp = pb.MakeRsp_GetUserInfoRsp(req.GetMid(), reply)
	}
	data, _ := proto.Marshal(rsp)
	c.send <- data
}

// handle notify
func (c *Client) Chat(ntf *pb.Notify) {
	chatNtf := ntf.GetChatNotify()
	msg := chatNtf.GetMessage()

	push := pb.MakePush_ChatPush(msg)

	data, _ := proto.Marshal(push)
	c.hub.broadcast <- []byte(data)
}
