package main

import (	
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	// "firebase.google.com/go/messaging"
)

type Menu struct {
	Connected bool
	Version string
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	m.Version = VERSION
	return `
<menu>
{{if .Connected}}
	<menuitem label="Connected to ProPresenter" disabled></menuitem>
	<menuitem separator></menuitem>
	<menuclocks>
{{else}}
	<menuitem label="Disconnected from ProPresenter" disabled></menuitem>
	<menuitem separator></menuitem>
	<instructions>
{{end}}
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

type MenuClocks struct {
	Clocks map[string]string
}

func (m *MenuClocks) Render() string {
	return `
	<menuitem label="Clocks: {{range $key, $value := .Clocks}}{{$value}} {{end}}" disabled></menuitem>
	<menuitem separator></menuitem>
`
}

func (m *MenuClocks) Subscribe() *app.EventSubscriber {
	return app.NewEventSubscriber().Subscribe("clock-event", m.OnClockChange)
}

func (m *MenuClocks) OnClockChange(clocks map[string]string) {
	m.Clocks = clocks
	app.Render(m)
}

type Instructions struct {
	Password string
	Port string
}
func (m *Instructions) Render() string {
	m.Password = WS_PASSWORD
	m.Port = WS_PORT
	return `
	<menuitem label="Setup: Open ProPresenter, go to Preferences, enable Network," disabled></menuitem>
	<menuitem label="set the Network Port to {{.Port}}. Enable Stage Display App, set" disabled></menuitem>
	<menuitem label="the password to {{.Password}} and ensure stage port is empty, restart PP." disabled></menuitem>
`
}

func main() {
	app.EnableDebug(true)

	app.Import(&Menu{})
	app.Import(&MenuClocks{})
	app.Import(&Instructions{})

	app.HandleAction("connect-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("connect-event", a.Arg)
	})
	app.HandleAction("clock-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("clock-event", a.Arg)
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
