package main

import (
	"encoding/json"
	"fmt"
	autostart "github.com/ProtonMail/go-autostart"
	"github.com/dchest/uniuri"
	"github.com/mitchellh/go-homedir"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/mac"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var LocationName string
var InstallUuid string
var AutostartEnabled bool
var autostartApp *autostart.App

func CheckAutostart() bool {
	myPath, _ := os.Executable()
	autostartApp = &autostart.App{
		Name:        "clockstream",
		DisplayName: "Stream ProPresenter clocks to hillsonglabs",
		Exec:        []string{myPath},
	}

	if autostartApp.IsEnabled() {
		fmt.Println("App is set to start on boot")
		return true
		// if err := app.Disable(); err != nil {
		// 	log.Fatal(err)
		// }
	} else {
		fmt.Println("App is NOT set to start on boot")
		return false
		// if err := app.Enable(); err != nil {
		// 	log.Fatal(err)
		// }
	}
}

func SetAutostart(will_autostart bool) {
	if will_autostart {
		if err := autostartApp.Enable(); err != nil {
			app.Log(err)
		} else {
			AutostartEnabled = true
		}
	} else {
		if err := autostartApp.Disable(); err != nil {
			app.Log(err)
		} else {
			AutostartEnabled = false
		}
	}
}

type Menu struct {
	Connected    bool
	Version      string
	LocationName string
	Autostart    bool
}

// Render returns the HTML describing the status menu.
func (m *Menu) Render() string {
	m.Version = VERSION
	m.LocationName = LocationName
	m.Autostart = AutostartEnabled
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
	<menuitem label="Location: {{.LocationName}}" disabled></menuitem>
	<menuitem label="Change..." onclick="OpenSettings"></menuitem>
	<menuitem label="Edit Config File..." onclick="OpenTextedit"></menuitem>
{{if .Autostart}}
	<menuitem label="Start on boot" onclick="ToggleAutostart" checked=true></menuitem>
{{else}}
	<menuitem label="Start on boot" onclick="ToggleAutostart"></menuitem>
{{end}}
	<menuitem separator></menuitem>
	<menuitem label="{{.Version}}" disabled></menuitem>
	<menuitem separator></menuitem>
	<menuitem label="Quit" selector="terminate:"></menuitem>
</menu>
	`
}

func (m *Menu) Subscribe() *app.EventSubscriber {
	return app.NewEventSubscriber().Subscribe("connect-event", m.OnConnectEvent).Subscribe("refresh-event", m.OnConnectEvent)
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

func (m *Menu) ToggleAutostart() {
	if AutostartEnabled {
		SetAutostart(false)
	} else {
		SetAutostart(true)
	}
	app.Render(m)
}

var SettingsWindow app.Window
var SettingsIsOpen bool

func (m *Menu) OpenSettings() {
	if SettingsIsOpen == false {
		SettingsIsOpen = true
		SettingsWindow = app.NewWindow(app.WindowConfig{
			Width:          450,
			Height:         300,
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
func (m *Menu) OpenTextedit() {
	err := open.Run(fmt.Sprintf("%s/.clockstream", os.Getenv("HOME"))) //open.Run("~/.clockstream")
	app.Logf("Command finished with error: %v", err)
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
<div class="Help" style="color:#555; padding:30px;">
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
	Error               bool
}

func (h *Settings) Render() string {
	if h.LocationInitialized == false {
		h.Location = LocationName
		h.LocationInitialized = true
		// TODO: How to initialize values? app.Import(&Settings{Location: LocationName})
	}
	return `
<div class="Help" style="color:#555; padding:30px;">
	<h3>Clockstream Settings</h3>
	<p>
		Location Name<br>
	</p>
	<p>
		<input value="{{.Location}}" style="font-family: mono;" placeholder="usa-east-(your-name)..." onchange="Location" autofocus style="font-size: 14pt; padding: 5px;" autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false">
		<br><small>Must be lowercase with hyphens, no spaces. <br>Example: &quot;usa-east-nyc&quot;)</small>
	</p>
	{{if .Error}}
	<p>Error: Invalid value. Ensure name is a lowercase alphanumeric with hyphens.</p>
	{{end}}
	<p>
		<input type="submit" value="Update" style="font-size: 14pt; padding: 5px;" autocomplete="off" onclick="OnUpdate">
	</p>
</div>
	`
}

func LowerDashify(str string) string {
	var re = regexp.MustCompile(`[^a-zA-Z\d-]`)
	return re.ReplaceAllString(strings.ToLower(str), "-")
}

func (h *Settings) OnUpdate() {
	lowerDashed := LowerDashify(h.Location)
	if h.Location != lowerDashed {
		h.Location = lowerDashed
		h.Error = true
		app.Render(h)
	} else {
		LocationName = lowerDashed
		app.PostAction("refresh-action", true)
		app.Log("Updated LocationName: ", LocationName)
		WriteConfig()
		SettingsWindow.Close()
	}
}

type Config struct {
	LocationName string `json:"location_name"`
	InstallUuid  string `json:"install_uuid"`
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
	InstallUuid = config.InstallUuid
	if LocationName == "" {
		WriteDefaultConfig()
	}
}

func WriteDefaultConfig() {
	generatedLocation := fmt.Sprintf("unknown-%s", uniuri.NewLenChars(5, []byte("abcdefghijklmnopqrstuvwxyz0123456789")))
	LocationName = generatedLocation
	InstallUuid = uniuri.NewLen(5)
	WriteConfig()
}

func WriteConfig() {
	if InstallUuid == "" {
		InstallUuid = uniuri.NewLen(5)
	}
	defaultConfig := []byte(fmt.Sprintf(`{"location_name":"%s", "install_uuid":"%s"}`, LocationName, InstallUuid))
	err := ioutil.WriteFile(ConfigPath(), defaultConfig, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	// app.EnableDebug(true)

	WriteConfig()

	LoadOrCreateConfig()
	AutostartEnabled = CheckAutostart()

	if InstallUuid == "" {
		// Upgrade v0.1 configs to v0.2
		WriteConfig()
	}

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
	app.HandleAction("refresh-action", func(e app.EventDispatcher, a app.Action) {
		e.Dispatch("refresh-event", a.Arg)
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
