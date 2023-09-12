package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{255, 255, 255, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

// This function writes a 128x128 white image with black text in the centre to disk
func getTestImage(text string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	addLabel(img, 20, 30, text)
	return img

}

// This function writes a the first frame of the gif to disk
func writeGifToFile(img *image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	options := gif.Options{}
	options.NumColors = 256

	return gif.Encode(f, *img, &options)
}

// This function creates a mygif and appends five images to it then writes to disk
func TestImageGeneration(t *testing.T) {
	img := getTestImage("hi")
	writeGifToFile(&img, "TestImageGeneration.gif")
}
func TestAppend(t *testing.T) {
	const frameCount = 10

	gif := initMyGIF(frameCount, 50, 10)

	for i := 0; i < frameCount*2; i++ {
		img := getTestImage(fmt.Sprint(i))
		gif.insertImage(&img, End)
	}
	gif.writeToFile("test.gif")

}
