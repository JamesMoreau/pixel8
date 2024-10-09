package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path"
	"strings"
	"syscall/js"

	"golang.org/x/image/draw"
)

/*
TODO
-add examples
-add custom color palette
-show the sample color palettes
-delete openImage
-see if i can move wasm to head tag
- whyy: originalColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
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
	fmt.Println("Hello web assembly from go!")
	js.Global().Set("generatePassword", js.FuncOf(jsWrapperPixel8))
	select {} // This runs forever so that the go program never exits.
}

func jsWrapperPixel8(this js.Value, inputs []js.Value) interface{} {
	return ""
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

	// Scale down the image to a smaller size, then scale back up to original size.
	scaledDownImg := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	draw.NearestNeighbor.Scale(scaledDownImg, scaledDownImg.Bounds(), img, img.Bounds(), draw.Over, nil)
	pixelatedImg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.NearestNeighbor.Scale(pixelatedImg, pixelatedImg.Bounds(), scaledDownImg, scaledDownImg.Bounds(), draw.Over, nil)

	return pixelatedImg
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
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
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

// Convert an image to grayscale using the "Average" method.
func convertToGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	gray := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			average := (r + g + b) / 3
			color := color.NRGBA64{R: uint16(average), G: uint16(average), B: uint16(average), A: uint16(a)}

			gray.Set(x, y, color)
		}
	}

	return gray
}