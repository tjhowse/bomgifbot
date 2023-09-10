package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v9"
)

type config struct {
	MastodonURL          string `env:"MASTODON_SERVER"`
	MastodonClientID     string `env:"MASTODON_CLIENT_ID"`
	MastodonClientSecret string `env:"MASTODON_CLIENT_SECRET"`
	MastodonUserEmail    string `env:"MASTODON_USER_EMAIL"`
	MastodonUserPassword string `env:"MASTODON_USER_PASSWORD"`
	ImageURL             string `env:"IMAGE_URL"`
	ImageUpdateInterval  int64  `env:"IMAGE_UPDATE_INTERVAL" envDefault:"300"`
	ImageFrameCount      int64  `env:"IMAGE_FRAME_COUNT" envDefault:"10"`
	ImageFrameDelay      int64  `env:"IMAGE_FRAME_DELAY" envDefault:"1"`
}

// This function downloads the image at the provided URL into the provided image.Image pointer.
func downloadImage(url string, img *image.Image) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	*img, err = gif.Decode(resp.Body)
	return err
}

func writeImageToFile(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// This inserts the provide image into the first frame of the gif,
// and shifts all the other frames down one.
func prependImageToGif(img *image.Image, gif *gif.GIF) error {
	// Shift all the frames down one.
	for i := len(gif.Image) - 1; i > 0; i-- {
		gif.Image[i] = gif.Image[i-1]
		gif.Delay[i] = gif.Delay[i-1]
	}
	// Palettise the image.Image into a image.Paletted
	palettedImage := image.NewPaletted((*img).Bounds(), nil)
	// Draw the image into the paletted image.
	draw.Draw(palettedImage, palettedImage.Bounds(), *img, (*img).Bounds().Min, draw.Src)
	// Insert the paletted image into the gif.
	gif.Image[0] = palettedImage
	return nil
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	var m *Mastodon
	var err error
	// Set the next update time to one second ago, so that the first update happens immediately.
	nextUpdateTime := time.Now().Add(-time.Second)
	// Somewhere to store the gif
	var gif gif.GIF
	// Initialise the gif's image buffer
	gif.Image = make([]*image.Paletted, cfg.ImageFrameCount)
	// Initialise the delays to the desired delay
	for i := range gif.Image {
		gif.Delay[i] = int(cfg.ImageFrameDelay)
	}

	for {
		// If the mastodon link is down, bring it back up.
		if m == nil {
			m, err = NewMastodon(cfg.MastodonURL, cfg.MastodonClientID, cfg.MastodonClientSecret)
			if err != nil {
				slog.Error(err.Error())
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
		err = downloadImage(cfg.ImageURL, &img)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		// Prepend the image to the gif.
		err = prependImageToGif(&img, &gif)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		// Upload the gif
		// Compose a toot with the gif attached (?)
		// Post the toot

	}
}
