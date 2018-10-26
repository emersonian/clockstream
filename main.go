package main

import (
	"os"
	"os/signal"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	"github.com/sacOO7/gowebsocket"
	// "firebase.google.com/go/messaging"
)

const URL = "ws://localhost:63968/stagedisplay"

func ConnectProPresenter(){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	socket := gowebsocket.New(URL)
  
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		app.Log("ws: Received connect error - ", err)
	}
  
	socket.OnConnected = func(socket gowebsocket.Socket) {
		app.Log("ws: Connected to server");
		app.PostAction("connect-action", true)
		// Log in to ProPresenter
		socket.SendText(`{"pwd":"h1llcl0cks","ptl":610,"acn":"ath"}`)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		app.Log("ws: Disconnected from server ")
		app.PostAction("connect-action", false)
		return
	}
  
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		app.Log("ws: Received message - " + message)
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

type Menu struct {
	Connected bool
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	return `
<menu>
	{{if .Connected}}
	<menuitem label="Connected to ProPresenter" disabled></menuitem>
	{{else}}
	<menuitem label="Disconnected from ProPresenter" disabled></menuitem>
	{{end}}
	<menuitem separator></menuitem>
	<menuitem label="Setup: Open ProPresenter, go to Preferences, Network, click" disabled></menuitem>
	<menuitem label="Enable Stage Display App. Set the password to h1llcl0cks and set" disabled></menuitem>
	<menuitem label="the Network Port (at top) to 63968 (leave stage port empty), restart PP." disabled></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Quit" selector="terminate:"></menuitem>
</menu>
	`
}

func (m *Menu) Subscribe() *app.EventSubscriber {
	return app.NewEventSubscriber().Subscribe("connect-event", m.OnConnectEvent)
}

func (m *Menu) OnConnectEvent(connected bool) {
	m.Connected = connected
	app.Render(m)
}

func main() {
	app.EnableDebug(true)
	app.Import(&Menu{})
	app.HandleAction("connect-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("connect-event", a.Arg)
	})
	app.Run(&mac.Driver{
		Bundle: mac.Bundle{
			Background: true,
		},

		OnRun: func() {
			app.NewStatusMenu(app.StatusMenuConfig{
				Icon: app.Resources("logo.png"),
				URL:  "/Menu",
			})
			ConnectProPresenter()
		},
	}, app.Logs())
}
