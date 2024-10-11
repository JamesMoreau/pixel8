package main

import (
	"bytes"
	"encoding/base64"
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

func main() {
	fmt.Println("Hello web assembly from Go!")
	js.Global().Set("processPixel8", js.FuncOf(jsWrapperProcessPixel8))
	select {} // This runs forever so that the Go program never exits.
}

func jsWrapperProcessPixel8(this js.Value, inputs []js.Value) interface{} {
	if len(inputs) < 3 {
		fmt.Println("Not enough arguments")
		return nil
	}

	imgData := inputs[0].String()
	img, err := decodeBase64Image(imgData)
	if err != nil {
		fmt.Println("Failed to decode image:", err)
		return nil
	}

	blockSize := inputs[1].Int()

	jsPalette := inputs[2]
	var palette []color.Color
	for i := 0; i < jsPalette.Length(); i++ {
		r := jsPalette.Index(i).Index(0).Int()
		g := jsPalette.Index(i).Index(1).Int()
		b := jsPalette.Index(i).Index(2).Int()
		a := jsPalette.Index(i).Index(3).Int()
		palette = append(palette, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
	}

	resultImg := processPixel8(img, blockSize, palette)
	resultBase64 := encodeImageToBase64(resultImg)

	return js.ValueOf(resultBase64)
}

func decodeBase64Image(data string) (image.Image, error) {
	data = data[strings.IndexByte(data, ',')+1:] // Remove the data URL prefix if it exists.

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func encodeImageToBase64(img image.Image) string {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		fmt.Println("Failed to encode image:", err)
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func processPixel8(img image.Image, blockSize int, palette []color.Color) image.Image {
	pixel8ed := pixel8(img, blockSize)

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
