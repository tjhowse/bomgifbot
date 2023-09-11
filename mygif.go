package main

import (
	"bufio"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
)

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

// This function writes the animated gif to a writer
func (g *myGIF) writeToWriter(w *bufio.Writer) error {
	return gif.EncodeAll(w, &g.GIF)
}

// This inserts the provide image into the first frame of the gif,
// and shifts all the other frames down one.
func (g *myGIF) prependImage(img *image.Image) error {
	if g.frameCount < g.maxFrameCount {
		g.frameCount++
		g.Image = append(g.Image, nil)
		g.Delay = append(g.Delay, int(g.frameDelay))
	}
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
