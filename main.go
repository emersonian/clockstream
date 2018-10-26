package main

import (	
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	// "firebase.google.com/go/messaging"
)

type Menu struct {
	Connected bool
	Clocks map[string]string
	Password string
	Port string
	Version string
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	m.Password = WS_PASSWORD
	m.Port = WS_PORT
	m.Version = VERSION
	return `
<menu>
	{{if .Connected}}
	<menuitem label="Connected to ProPresenter" disabled></menuitem>
	{{else}}
	<menuitem label="Disconnected from ProPresenter" disabled></menuitem>
	{{end}}
	<menuitem separator></menuitem>
	{{range .Clocks}}
	<menuitem label="Clock: {{.}}"></menuitem>
	{{end}}
	<menuitem label="Setup: Open ProPresenter, go to Preferences, enable Network," disabled></menuitem>
	<menuitem label="set the Network Port to {{.Port}}. Enable Stage Display App, set" disabled></menuitem>
	<menuitem label="the password to {{.Password}} and ensure stage port is empty, restart PP." disabled></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="{{.Version}}" disabled></menuitem>
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
