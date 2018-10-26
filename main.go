package main

import (
	"log"
	"time"
	"net/http"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	"github.com/gorilla/websocket"
)

func ConnectProPresenter(){
	// interrupt := make(chan os.Signal, 1)
	url := "ws://localhost:63968/stagedisplay"
	log.Printf("connecting to %s", url)

	var Dialer = &websocket.Dialer{
	    Proxy:            http.ProxyFromEnvironment,
	    HandshakeTimeout: 5 * time.Second,
	}

	c, _, err := Dialer.Dial(url, nil)
	if err == nil {
		log.Println("connected")
	    if err := c.WriteMessage(websocket.TextMessage, []byte(`{"pwd":"h1llcl0cks","ptl":610,"acn":"ath"}`)); err != nil {
	        log.Println("auth:", err)
	        return
	    }
	} else {
		log.Println("dialerr:", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("readerr:", err)
				return
			}
			log.Printf("recv: %s", message)
			// if message == `{"acn":"ath","ath":true,"err":""}` {
			// 	log.Println("auth:success")
			// }
			
		}
	}()
}

func main() {
	app.Import(&Menu{})
	app.Run(&mac.Driver{
		Bundle: mac.Bundle{
			Background: true,
		},

		OnRun: func() {
			app.NewStatusMenu(app.StatusMenuConfig{
				Icon: app.Resources("logo.png"),
				// Text: "Background app",
				URL:  "/Menu",
			})
		},
	}, app.Logs())

	ConnectProPresenter()
}

// Menu is a component that describes a status able to change its text and icon.
type Menu struct {
	IconHidden bool
	TextHidden bool
	Status string
	Connected bool
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	m.Status = "Disconnected"
	// m.Status = "Connected to clocks.hillsonglabs.com"
	return `
<menu>
	<menuitem label="{{.Status}}"></menuitem>
	{{if not .Connected}}
	<menuitem separator></menuitem>
	<menuitem label="Setup: Open ProPresenter, go to Preferences, Network, click" disabled></menuitem>
	<menuitem label="Enable Stage Display App, then set the password to h1llcl0cks and" disabled></menuitem>
	<menuitem label="set the Network Port (at top) to 63968 (leave stage port empty), restart ProPresenter." disabled></menuitem>
	{{end}}
	<menuitem separator></menuitem>
	<menuitem label="Debug" onclick="OnOpenWindow"></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Quit" selector="terminate:"></menuitem>
</menu>
	`
	// 
	// 	{{if .Connected}}
	// <menuitem label="Disconnect" onclick="OnHideText"></menuitem>
	// {{else}}
	// <menuitem label="Connect" onclick="OnShowText"></menuitem>
	// {{end}}
}

// OnShowIcon is the function called when the show icon button is clicked.
func (m *Menu) OnShowIcon() {
	app.ElemByCompo(m).WhenStatusMenu(func(s app.StatusMenu) {
		s.SetIcon(app.Resources("logo.png"))
		m.IconHidden = false
		app.Render(m)
	})
}

// OnHideIcon is the function called when the hide icon button is clicked.
func (m *Menu) OnHideIcon() {
	app.ElemByCompo(m).WhenStatusMenu(func(s app.StatusMenu) {
		s.SetIcon("")
		m.IconHidden = true
		app.Render(m)

		if m.TextHidden {
			m.OnShowText()
		}
	})

}

// OnShowText is the function called when the show text button is clicked.
func (m *Menu) OnShowText() {
	app.ElemByCompo(m).WhenStatusMenu(func(s app.StatusMenu) {
		s.SetText("Background app")
		m.TextHidden = false
		app.Render(m)
	})
}

// OnHideText is the function called when the hide text button is clicked.
func (m *Menu) OnHideText() {
	app.ElemByCompo(m).WhenStatusMenu(func(s app.StatusMenu) {
		s.SetText("")
		m.TextHidden = true
		app.Render(m)

		if m.IconHidden {
			m.OnShowIcon()
		}
	})
}

// OnOpenWindow is the function called when the open a window button is clicked.
func (m *Menu) OnOpenWindow() {
	app.NewWindow(app.WindowConfig{
		X: 0,
		Y: 100,
		Width:          400,
		Height:         300,
		TitlebarHidden: true,
	})
}
