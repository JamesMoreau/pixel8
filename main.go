package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path"
	"strings"
)

/*
TODO
-should i always start for loops from 0 if img rectangle doesn't necessarily start at 0.
*/

var (
	nesPalette = []color.Color {
		color.RGBA{0x80, 0x80, 0x80, 0xFF}, // Gray
		color.RGBA{0x00, 0x00, 0xFF, 0xFF}, // Blue
		color.RGBA{0x00, 0xFF, 0x00, 0xFF}, // Green
		color.RGBA{0xFF, 0x00, 0x00, 0xFF}, // Red
		color.RGBA{0xFF, 0xFF, 0x00, 0xFF}, // Yellow
		color.RGBA{0xFF, 0xA5, 0x00, 0xFF}, // Orange
		color.RGBA{0x00, 0x00, 0x00, 0xFF}, // Black
		color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, // White
	}
)

func main() {
	img, err := openImage("assets/asuka.jpeg")
	if err != nil {
		log.Fatal(err)
	}

	pixel8ed := processPixel8(img, 8, false, nesPalette)

	err = saveImageToFile(pixel8ed, "output/asuka.jpg")
	if err != nil {
		log.Fatal(err)
	}
}

func processPixel8(img image.Image, blockSize int, grayscale bool, palette []color.Color) image.Image {
	pixel8ed := pixel8(img, blockSize)
	
	if grayscale {
		return convertToGrayscale(pixel8ed)
	}
	
	usePalette := palette != nil
	if usePalette {
		return convertToColorPalette(pixel8ed, palette)
	}

	return pixel8ed
}

func openImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil { 
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func pixel8(img image.Image, blockSize int) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	scaledW := int(math.Ceil(float64(width) / float64(blockSize)))
	scaledH := int(math.Ceil(float64(height) / float64(blockSize)))

	// Resize the image to scaled down size, then resize again to original size.
	scaledDownImg := resizeWithNearestNeighbour(img, scaledW, scaledH)
	pixelatedImg := resizeWithNearestNeighbour(scaledDownImg, width, height)

	return pixelatedImg
}

func resizeWithNearestNeighbour(img image.Image, newWidth, newHeight int) image.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	oldWidth := img.Bounds().Dx()
	oldHeight := img.Bounds().Dy()

	xScale := float64(oldWidth) / float64(newWidth)
	yScale := float64(oldHeight) / float64(newHeight)

	for x := 0; x < newWidth; x++ {
		for y := 0; y < newHeight; y++ {
			// Find the nearest pixel in the original image
			srcX := int(float64(x) * xScale)
			srcY := int(float64(y) * yScale)

			nearestColor := img.At(srcX, srcY)
			newImg.Set(x, y, nearestColor)
		}
	}

	return newImg
}

func saveImageToFile(img image.Image, filepath string) error {
	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	ext := strings.ToLower(path.Ext(filepath))
	switch ext {
	case ".png":
		err = png.Encode(outFile, img)
		if err != nil {
			return fmt.Errorf("failed to encode PNG: %v", err)
		}
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 100})
		if err != nil {
			return fmt.Errorf("failed to encode JPEG: %v", err)
		}
	default:
		return fmt.Errorf("unsupported file extension: %v", ext)
	}

	return nil
}

func convertToColorPalette(img image.Image, palette []color.Color) image.Image {
	bounds := img.Bounds()
	paletteImg := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			closestColor := findClosestPaletteColor(originalColor, palette)

			paletteImg.Set(x, y, closestColor)
		}
	}

	return paletteImg
}

func findClosestPaletteColor(c color.Color, palette []color.Color) color.Color {
	closestColor := palette[0]
	minDistance := math.MaxFloat64

	for _, p := range palette {
		pr, pg, pb, _ := p.RGBA()
		cr, cg, cb, _ := c.RGBA()
		
		dr := float64(cr>>8) - float64(pr>>8)
		dg := float64(cg>>8) - float64(pg>>8)
		db := float64(cb>>8) - float64(pb>>8)

		distance := dr*dr + dg*dg + db*db
		if distance < minDistance {
			minDistance = distance
			closestColor = p
		}
	}

	return closestColor
}

// This grayscale converter uses the "Average" method.
func convertToGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	gray := image.NewRGBA(bounds)


	for x := bounds.Dx(); x < bounds.Dx(); x++ {

	}

	return gray
}