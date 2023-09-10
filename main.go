package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v9"
)

type config struct {
	MastodonURL          string `env:"MASTODON_SERVER"`
	MastodonClientID     string `env:"MASTODON_CLIENT_ID"`
	MastodonClientSecret string `env:"MASTODON_CLIENT_SECRET"`
	MastodonUserEmail    string `env:"MASTODON_USER_EMAIL"`
	MastodonUserPassword string `env:"MASTODON_USER_PASSWORD"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	var m *Mastodon
	var err error

	for {
		if m == nil {
			m, err = NewMastodon(cfg.MastodonURL, cfg.MastodonClientID, cfg.MastodonClientSecret)
			if err != nil {
				slog.Error(err.Error())
				time.Sleep(10 * time.Second)
				continue
			}
		}
		m.PostStatus("Hello, world!")
		time.Sleep(5 * time.Minute)
	}
}
