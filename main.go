package main

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/mitchellh/go-homedir"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	"io/ioutil"
	"os"
)

var LocationName string

type Menu struct {
	Connected bool
	Version   string
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
	<menuitem label="Setup Instructions" onclick="OpenSetup"></menuitem>
	<menuitem separator></menuitem>
{{end}}
	<menuitem separator></menuitem>
	<menuitem label="Set location name..." onclick="OpenSettings"></menuitem>
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

func (m *Menu) OpenSetup() {
	app.NewWindow(app.WindowConfig{
		Width:          450,
		Height:         250,
		X:              400,
		Y:              600,
		TitlebarHidden: true,
		FixedSize:      true,
		URL:            "help",
	})
}

var SettingsWindow app.Window
var SettingsIsOpen bool

func (m *Menu) OpenSettings() {
	if SettingsIsOpen == false {
		SettingsIsOpen = true
		SettingsWindow = app.NewWindow(app.WindowConfig{
			Width:          450,
			Height:         250,
			X:              400,
			Y:              600,
			TitlebarHidden: true,
			FixedSize:      true,
			URL:            "settings",
			OnClose: func() bool {
				SettingsIsOpen = false
				return true
			},
		})
	}
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
	Port     string
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

type Settings struct {
	Location            string
	LocationInitialized bool
}

func (h *Settings) Render() string {
	if h.LocationInitialized == false {
		h.Location = LocationName
		h.LocationInitialized = true
		// TODO: How to initialize values? app.Import(&Settings{Location: LocationName})
	}
	return `
<div class="Help" style="color:#ddd; padding:30px;">
	<h3>Clockstream Settings</h3>
	<p>
		Location Name<br>
		<small>(alphanumeric, no spaces. Example: &quot;usa-east-nyc&quot;)</small>
	</p>
	<p>
		<input value="{{.Location}}" placeholder="usa-east-(your-name)..." onchange="Location" autofocus style="font-size: 14pt; padding: 5px;" autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false">
	</p>
	<p>

		<input type="submit" value="Update" style="font-size: 14pt; padding: 5px;" autocomplete="off" onclick="OnUpdate">
	</p>
</div>
	`
}

func (h *Settings) OnUpdate() {
	LocationName = h.Location
	WriteConfig()
	app.Log("Updated LocationName: ", LocationName)
	SettingsWindow.Close()
}

type Config struct {
	LocationName string `json:"location_name"`
}

const CONFIG_PATH = "/.clockstream"

func ConfigPath() string {
	str, _ := homedir.Expand(fmt.Sprintf("~/%s", CONFIG_PATH))
	return str
}

func LoadOrCreateConfig() {
	var config Config
	// Read the config. If location_name is empty or not exists, overwrite the file.
	configFile, err := os.Open(ConfigPath())
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Writing default config file.")
		WriteDefaultConfig()
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	LocationName = config.LocationName
	if LocationName == "" {
		WriteDefaultConfig()
	}
}

func WriteDefaultConfig() {
	generatedLocation := fmt.Sprintf("unknown-%s", uniuri.NewLenChars(5, []byte("abcdefghijklmnopqrstuvwxyz0123456789")))
	LocationName = generatedLocation
	WriteConfig()
}

func WriteConfig() {
	defaultConfig := []byte(fmt.Sprintf(`{"location_name":"%s"}`, LocationName))
	err := ioutil.WriteFile(ConfigPath(), defaultConfig, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	app.EnableDebug(true)

	LoadOrCreateConfig()

	app.Log("LocationName: ", LocationName)

	app.Import(&Menu{})
	app.Import(&MenuClocks{})
	app.Import(&Help{})
	app.Import(&Settings{})

	app.HandleAction("connect-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("connect-event", a.Arg)
	})
	app.HandleAction("clock-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("clock-event", a.Arg)
	})
	app.Log(LocationName)

	app.Run(&mac.Driver{
		Bundle: mac.Bundle{
			Background: true,
		},

		OnRun: func() {
			app.NewStatusMenu(app.StatusMenuConfig{
				Icon: app.Resources("logo.png"),
				URL:  "/Menu",
			})
			ConnectionWatchdog()
		},
	}, app.Logs())
}
