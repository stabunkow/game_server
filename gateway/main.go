package main

import (
	"fmt"
	"game_server/model"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func init() {
	GetHub()
	fmt.Println("Gateway Server Start ...")
}

func main() {
	// handle user login
	http.HandleFunc("/api/login", login)
	http.HandleFunc("/api/register", register)
	http.HandleFunc("/ws", serveWs)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	sid := r.FormValue("token")
	if sid == "" {
		closeWs(conn, "no token")
		return
	}

	usr := model.FindUserBySid(sid)
	if usr == nil {
		closeWs(conn, "token invaild")
		return
	}

	uid := usr.GetId()
	// avoid multiple login
	client := GetHub().GetClient(uid)
	if client != nil {
		client.ExitWithReason("multiple login")
	}

	if !model.UserStorageExists(uid) {
		usr.Storage()
	}

	newClient(conn, GetHub(), uid)
}

func closeWs(conn *websocket.Conn, reason string) {
	defer conn.Close()
	err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason), time.Now().Add(time.Second))
	if err != nil {
		return
	}
}
