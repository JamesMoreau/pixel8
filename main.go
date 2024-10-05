package main

import (
	"fmt"
	"image"
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
-Grayscale
-Color Palettes
*/

func main() {
	img, err := openImage("assets/asuka.jpeg")
	if err != nil {
		log.Fatal(err)
	}

	blockSize := 16
	pixelated := pixel8(img, blockSize)
	err = saveImageToFile(pixelated, "output/asuka.jpg")
	if err != nil {
		log.Fatal(err)
	}
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
