package main

import (
	"encoding/json"
	"fmt"
	"github.com/murlokswarm/app"
	"github.com/sacOO7/gowebsocket"
	"os"
	"os/signal"
	"time"
)

var clocks = make(map[string]string)
var ProPresenterConnected bool
var ProPresenterConnecting bool
var LastPingAt time.Time

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
		app.PostAction("clock-action", clocks)
	}
	PushToFirebase(LocationClock{
		// InstallUuid: nil,
		Location:  LocationName,
		ClockUuid: message.Uid,
		Value:     message.Txt,
	})
}

var closeSocket chan bool

func ConnectProPresenter() {
	ProPresenterConnecting = true
	if closeSocket == nil {
		closeSocket = make(chan bool, 1)
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// TODO: Ensure connection timeout is short, no concurrent watchdog connection attempts
	socket := gowebsocket.New(WS_URL)

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		app.Log("ws: Received connect error - ", err)
		ProPresenterConnecting = false
		ProPresenterConnected = false
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		ProPresenterConnecting = false
		ProPresenterConnected = true
		app.Log("ws: Connected to server")
		app.PostAction("connect-action", true)
		// Log in to ProPresenter
		socket.SendText(fmt.Sprintf(`{"pwd":"%s","ptl":610,"acn":"ath"}`, WS_PASSWORD))
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		ProPresenterConnecting = false
		ProPresenterConnected = false
		app.Log("ws: Disconnected from server ")
		app.PostAction("connect-action", false)
		return
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		app.Log("ws: Received message - " + message)
		HandleMessage(message)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		LastPingAt = time.Now()
		app.Log("ws: Received ping - " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		app.Log("ws: Received pong - " + data)
	}

	socket.Connect()

	go func() {
		for {
			select {
			case <-closeSocket:
				app.Log("Socket: Closing.")
				socket.Close()
				return
			case <-interrupt:
				app.Log("Socket: Interrupt.")
				socket.Close()
				return
			}
		}
	}()
}

func ConnectionWatchdog() {
	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for _ = range ticker.C {
			if ProPresenterConnected {
				if time.Now().Sub(LastPingAt).Seconds() > 10 {
					closeSocket <- true
					ProPresenterConnecting = false
					ProPresenterConnected = false
					app.Log("Watchdog: Ping timeout, marked as disconnected.")
				}
			} else {
				if ProPresenterConnecting == false {
					app.Log("Watchdog: spawning connection attempt")
					ConnectProPresenter()
				}
			}
		}
	}()
}
