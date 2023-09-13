package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"math"
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

func FastCompare(img1, img2 *image.RGBA) (int64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := int64(0)

	for i := 0; i < len(img1.Pix); i++ {
		accumError += int64(sqDiffUInt8(img1.Pix[i], img2.Pix[i]))
	}

	return int64(math.Sqrt(float64(accumError))), nil
}

func sqDiffUInt8(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
}

// This function creates a mygif and appends five images to it then writes to disk
func TestImageGeneration(t *testing.T) {
	img := getTestImage("hi")
	writeGifToFile(&img, "TestImageGeneration.gif")
}
func TestAppend(t *testing.T) {
	const frameCount = 10
	testDir := t.TempDir()

	var originals [frameCount * 2]image.Image

	testGif := initMyGIF(frameCount, 50, 10)

	for i := 0; i < frameCount*2; i++ {
		img := getTestImage(fmt.Sprint(i))
		originals[i] = img
		testGif.insertImage(&img, End)
	}
	testGif.writeToFile(testDir + "/test.gif")

	// Load the gif back in
	f, err := os.Open(testDir + "/test.gif")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	gif2, err := gif.DecodeAll(f)
	if err != nil {
		t.Error(err)
	}

	// Step through each frame in the imported gif and compare it to the last ten frames
	// of "original" slice.

	for i := 0; i < frameCount; i++ {
		// Compare the two images
		// Convert gif2.Image[i] to RGBA
		gif2Image := image.NewRGBA(gif2.Image[i].Bounds())
		for x := 0; x < gif2.Image[i].Bounds().Max.X; x++ {
			for y := 0; y < gif2.Image[i].Bounds().Max.Y; y++ {
				gif2Image.Set(x, y, gif2.Image[i].At(x, y))
			}
		}
		accumError, err := FastCompare(originals[frameCount+i].(*image.RGBA), gif2Image)
		if err != nil {
			t.Error(err)
		}
		// Magic threshold number
		if accumError > 32615 {
			t.Errorf("Frame %d: %d", i, accumError)
		}
	}

}
