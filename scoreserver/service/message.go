package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/zer0pts/zer0ptsctfd/scoreserver/model"
)

const (
	sendDeadline = 10 * time.Second
	tickInterval = 10 * time.Second
	redisChannel = "message"
)

type message struct {
	Body          []byte `json:"body"`
	LoginRequired bool   `json:"login"`
	AdminRequired bool   `json:"admin"`
}

// ----

type MessageApp interface {
	HandleMessage() error
	Send(msg []byte, loginRequired, adminRequired bool)
	Add(c MessageClient)
	Remove(c MessageClient)

	NewClient(ws *websocket.Conn, user *model.User) MessageClient
}

type messageApp struct {
	msg     chan message
	clients map[MessageClient]struct{}
	add     chan MessageClient
	remove  chan MessageClient
}

// ----

type MessageClient interface {
	User() *model.User
	Send(msg []byte)
	SendHandler()
	Close()
}

type messageClient struct {
	msg     chan []byte
	service MessageApp
	ws      *websocket.Conn
	user    *model.User
}

func (c *messageClient) User() *model.User {
	return c.user
}

func (c *messageClient) Send(msg []byte) {
	c.msg <- msg
}

func (c *messageClient) SendHandler() {
	defer c.Close()

	t := time.NewTicker(tickInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			err := c.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(sendDeadline))
			if err != nil {
				return
			}

		case m := <-c.msg:
			c.ws.SetWriteDeadline(time.Now().Add(sendDeadline))
			if err := c.ws.WriteMessage(websocket.TextMessage, m); err != nil {
				return
			}
		}
	}
}

func (c *messageClient) Close() {
	c.service.Remove(c)
	c.ws.Close()
	close(c.msg)
}

// ----

func newMessageApp() *messageApp {
	return &messageApp{
		msg:     make(chan message),
		clients: make(map[MessageClient]struct{}),
		add:     make(chan MessageClient),
		remove:  make(chan MessageClient),
	}
}

func (app *app) NewClient(ws *websocket.Conn, user *model.User) MessageClient {
	return &messageClient{
		service: app,
		ws:      ws,
		user:    user,
		msg:     make(chan []byte),
	}
}

func (app *app) HandleMessage() error {
	pubsub := app.redis.Subscribe(redisChannel)
	sub := pubsub.Channel()
	for {
		select {
		case m := <-app.msg:
			payload, err := json.Marshal(m)
			if err != nil {
				log.Println(err)
				break
			}
			app.redis.Publish(redisChannel, string(payload))

		case m := <-sub:
			var msg message
			err := json.Unmarshal([]byte(m.Payload), &msg)
			if err != nil {
				log.Println(err)
				break
			}
			for c, _ := range app.clients {
				user := c.User()
				if msg.LoginRequired && user == nil {
					continue
				}
				if msg.AdminRequired && (user == nil || user.IsAdmin == false) {
					continue
				}
				c.Send(msg.Body)
			}

		case c := <-app.add:
			app.clients[c] = struct{}{}

		case c := <-app.remove:
			delete(app.clients, c)
		}
	}
	return nil
}
func (app *app) Send(msg []byte, loginRequired, adminRequired bool) {
	app.msg <- message{
		Body:          msg,
		LoginRequired: loginRequired,
		AdminRequired: adminRequired,
	}
}

func (app *app) Add(c MessageClient) {
	app.add <- c
}
func (app *app) Remove(c MessageClient) {
	app.remove <- c
}
