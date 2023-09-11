package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/jlaffaye/ftp"
)

type config struct {
	MastodonURL          string `env:"MASTODON_SERVER"`
	MastodonClientID     string `env:"MASTODON_CLIENT_ID"`
	MastodonClientSecret string `env:"MASTODON_CLIENT_SECRET"`
	MastodonUserEmail    string `env:"MASTODON_USER_EMAIL"`
	MastodonUserPassword string `env:"MASTODON_USER_PASSWORD"`
	ImageURL             string `env:"IMAGE_URL"`
	ImageURLParsed       *url.URL
	ImageUpdateInterval  int64 `env:"IMAGE_UPDATE_INTERVAL" envDefault:"300"`
	ImageFrameCount      int64 `env:"IMAGE_FRAME_COUNT" envDefault:"10"`
	ImageFrameDelay      int64 `env:"IMAGE_FRAME_DELAY" envDefault:"1"`
}

// This function downloads the image at the provided http/s URL into the provided image.Image pointer.
func downloadImageHTTP(url *url.URL, img *image.Image) error {
	resp, err := http.Get(url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	*img, err = gif.Decode(resp.Body)
	return err
}

// This function downloads the image at the provided FTP URL into the provided image.Image pointer.
func downloadImageFTP(url *url.URL, img *image.Image) error {
	c, err := ftp.Dial(url.Host+":21", ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return err
	}

	err = c.Login("anonymous", "anonymous")
	if err != nil {
		return err
	}

	r, err := c.Retr(url.Path)
	if err != nil {
		return err
	}
	defer r.Close()

	*img, err = gif.Decode(r)
	if err != nil {
		return err
	}

	if err := c.Quit(); err != nil {
		return err
	}
	return nil
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	// Parse the image URL
	var err error
	var m *Mastodon
	var b bytes.Buffer

	cfg.ImageURLParsed, err = url.Parse(cfg.ImageURL)
	if err != nil {
		slog.Error("Failed to parse image URL: " + err.Error())
		os.Exit(1)
	}

	// Set the next update time to one second ago, so that the first update happens immediately.
	nextUpdateTime := time.Now().Add(-time.Second)
	// Somewhere to store the gif
	gif := initMyGIF(cfg.ImageFrameCount, cfg.ImageFrameDelay)

	for {
		// If the mastodon link is down, bring it back up.
		if m == nil {
			m, err = NewMastodon(cfg.MastodonURL, cfg.MastodonClientID, cfg.MastodonClientSecret)
			if err != nil {
				slog.Error("Failed to connect to mastodon: " + err.Error())
				time.Sleep(10 * time.Second)
				continue
			}
		}
		// If it's not time to update yet, sleep a second.
		if time.Now().Before(nextUpdateTime) {
			time.Sleep(1 * time.Second)
			continue
		}
		// Calculate the next update time.
		nextUpdateTime = time.Now().Add(time.Duration(cfg.ImageUpdateInterval) * time.Second)

		// Download the image.
		var img image.Image

		switch cfg.ImageURLParsed.Scheme {
		case "ftp":
			err = downloadImageFTP(cfg.ImageURLParsed, &img)
		case "http":
			fallthrough
		case "https":
			err = downloadImageHTTP(cfg.ImageURLParsed, &img)
		default:
			slog.Error("Unrecognised URL scheme: " + cfg.ImageURLParsed.Scheme)
			os.Exit(1)
		}

		if err != nil {
			slog.Error("Failed to download and parse image: " + err.Error())
			time.Sleep(10 * time.Second)
			continue
		}

		// Prepend the image to the gif.
		err = gif.prependImage(&img)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// Write the gif to a buffer
		err = gif.writeToWriter(bufio.NewWriter(&b))
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// Post the gif to mastodon
		err = m.PostStatusWithImageFromReader("A gif, just for you.", bytes.NewReader(b.Bytes()))
		if err != nil {
			slog.Error(err.Error())
			m = nil
		}
		b.Reset()
	}
}
