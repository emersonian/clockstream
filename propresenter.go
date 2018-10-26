package main

import (
	"os"
	"os/signal"
	"fmt"
	"encoding/json"
	"github.com/murlokswarm/app"
	"github.com/sacOO7/gowebsocket"
)

var clocks = make(map[string]string)

func HandleMessage(str string) {
	type Message struct {
		Acn string
		Uid string
		Txt string
	}
	message := Message{}
	err := json.Unmarshal([]byte(str), &message)
	if err != nil {
		app.Log("JSON parsing error:", err)
	}
	if message.Acn == "tmr" {
		clocks[message.Uid] = message.Txt
		app.Log("Current clock: ", message.Txt)
		app.PostAction("refresh-clocks", true)
	}
}

func ConnectProPresenter(){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New(WS_URL)
  
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		app.Log("ws: Received connect error - ", err)
	}
  
	socket.OnConnected = func(socket gowebsocket.Socket) {
		app.Log("ws: Connected to server");
		app.PostAction("connect-action", true)
		// Log in to ProPresenter
		socket.SendText(fmt.Sprintf(`{"pwd":"%s","ptl":610,"acn":"ath"}`, WS_PASSWORD))
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		app.Log("ws: Disconnected from server ")
		app.PostAction("connect-action", false)
		return
	}
  
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		app.Log("ws: Received message - " + message)
		HandleMessage(message)
	}
  
  	// TODO: Should receive Ping every 10 seconds. Reconnect if no ping within that.
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		app.Log("ws: Received ping - " + data)
	}
  
    socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		app.Log("ws: Received pong - " + data)
	}
  
	socket.Connect()

  	go func() {
		for {
			select {
			case <-interrupt:
				app.Log("Interrupt. Closing websocket cleanly.")
				socket.Close()
				return
			}
		}
	}()
}
