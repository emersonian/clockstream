package main

import (
	firebase "firebase.google.com/go"
	firedb "firebase.google.com/go/db"
	"fmt"
	"github.com/murlokswarm/app"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"time"
)

type LocationClock struct {
	Location    string
	InstallUuid string
	ClockUuid   string
	Value       string
	UpdatedAt   time.Time
}

var Client *firedb.Client

func init() {
	ctx := context.Background()
	ao := map[string]interface{}{"uid": "clockstream"}
	conf := &firebase.Config{
		DatabaseURL:  FIRE_URL,
		AuthOverride: &ao,
	}

	opt := option.WithCredentialsJSON([]byte(GOOGLE_JSON_KEY))
	fbApp, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		app.Log("Firebase: Error initializing app:", err)
	}

	Client, err = fbApp.Database(ctx)
	if err != nil {
		app.Log("Firebase: Error initializing database client:", err)
	}
}

func PushToFirebase(clock LocationClock) {
	if Client == nil {
		return
	}

	ctx := context.Background()

	if clock.ClockUuid == "" {
		clock.ClockUuid = "time"
	}

	value := map[string]interface{}{
		"value": clock.Value,
	}

	clockKey := fmt.Sprintf("locations/%s/%s", LocationName, clock.ClockUuid)
	if err := Client.NewRef(clockKey).Update(ctx, value); err != nil {
		app.Log(err)
	}

	locationKey := fmt.Sprintf("locations/%s", LocationName)
	if err := Client.NewRef(locationKey).Update(ctx, map[string]interface{}{
		"install_uuid": clock.InstallUuid,
		"updated_at":   time.Now(),
	}); err != nil {
		app.Log(err)
	}
}
