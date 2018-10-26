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
	<menuitem label="Setup Instructions" onclick="OnOpenWindow"></menuitem>
	<menuitem separator></menuitem>
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

func (m *Menu) OnOpenWindow() {
	app.NewWindow(app.WindowConfig{
		Width:          450,
		Height:         250,
		X: 400,
		Y: 600,
		TitlebarHidden: true,
		URL: "help",
	})
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

type Help struct {
	Password string
	Port string
}

func (h *Help) Render() string {
	h.Password = WS_PASSWORD
	h.Port = WS_PORT
	return `
<div class="Help" style="color:#ddd; padding:30px;">
	<h3>Clockstream Setup Instructions</h3>
	<p>
		Open ProPresenter, go to Preferences, enable Network, set the Network
		Port to {{.Port}}.
	</p>
	<p>
		Enable Stage Display App, set the password to {{.Password}}
		and ensure stage port is empty, restart ProPresenter.
	</p>
	<p>
		You should see clocks when you click on clockstream's menu icon.
		Restart clockstream if it is still not working.
	</p>
</div>
	`
}


func main() {
	app.EnableDebug(true)

	app.Import(&Menu{})
	app.Import(&MenuClocks{})
	app.Import(&Help{})

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
