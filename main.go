package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
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

type myGIF struct {
	gif.GIF
	frameDelay    int64
	frameCount    int64
	maxFrameCount int64
}

func initMyGIF(maxFrameCount int64, frameDelay int64) *myGIF {
	g := myGIF{}
	g.Image = make([]*image.Paletted, 0)
	g.Delay = make([]int, 0)
	g.frameCount = 0
	g.frameDelay = frameDelay
	g.maxFrameCount = maxFrameCount
	return &g
}

// This function writes a the first frame of the gif to disk
func (g *myGIF) writeFirstFrameToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	options := gif.Options{}
	options.NumColors = 256

	return gif.Encode(f, g.Image[0], &options)
}

// This function writes an animated gif to disk.
func (g *myGIF) writeToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, &g.GIF)
}

// This inserts the provide image into the first frame of the gif,
// and shifts all the other frames down one.
func (g *myGIF) prependImage(img *image.Image) error {
	if g.frameCount < g.maxFrameCount {
		g.frameCount++
		g.Image = append(g.Image, nil)
		g.Delay = append(g.Delay, int(g.frameDelay))
	}
	slog.Info("Prepended image, up to frame " + fmt.Sprintf("%d", g.frameCount))
	// Shift all the frames down one.
	for i := len(g.Image) - 1; i > 0; i-- {
		g.Image[i] = g.Image[i-1]
		g.Delay[i] = g.Delay[i-1]
	}
	// Palettise the image.Image into a image.Paletted
	palettedImage := image.NewPaletted((*img).Bounds(), palette.Plan9)
	// Draw the image into the paletted image.
	draw.Draw(palettedImage, palettedImage.Bounds(), *img, (*img).Bounds().Min, draw.Src)
	// Insert the paletted image into the gif.
	g.Image[0] = palettedImage
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
			continue
		}
		// Prepend the image to the gif.
		err = gif.prependImage(&img)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// Write a single frame  to disk to test
		// err = gif.writeFirstFrameToFile("test.gif")
		// if err != nil {
		// 	slog.Error(err.Error())
		// 	continue
		// }
		// Write a single frame  to disk to test
		err = gif.writeToFile("test.gif")
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		// Upload the gif
		// Compose a toot with the gif attached (?)
		// Post the toot

	}
}
