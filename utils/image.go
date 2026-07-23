package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"strconv"

	"golang.org/x/image/draw"
)

func clamp(value, low, high int) int {
	if value < low {
		return low
	}

	if value > high {
		return high
	}

	return value
}

func HasAlpha(img image.Image) bool {
	switch img.(type) {
	case *image.NRGBA, *image.NRGBA64, *image.RGBA, *image.RGBA64:
		return true
	default:
		return false
	}
}

func parseWidth(width string) int {
	w := 256

	num, err := strconv.Atoi(width)
	if err != nil {
		num = w
	}

	return clamp(num, 32, 512)

}

func ResizeImage(data []byte, targetWidth string) (image.Image, error) {
	width := parseWidth(targetWidth)

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return img, nil
	}

	if srcWidth <= width {
		return img, nil
	}

	targetHeight := max(int(float64(srcHeight)*float64(width)/float64(srcWidth)), 1)

	dest := image.NewRGBA(image.Rect(0, 0, width, targetHeight))
	draw.CatmullRom.Scale(dest, dest.Bounds(), img, bounds, draw.Over, nil)

	return dest, nil
}

func ResizeAndEncode(data []byte, targetWidth string) ([]byte, string, error) {
	img, err := ResizeImage(data, targetWidth)
	if err != nil {
		return nil, "", err
	}

	var out bytes.Buffer
	var contentType string

	if HasAlpha(img) {
		contentType = "image/png"
		err = png.Encode(&out, img)
	} else {
		contentType = "image/jpeg"
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: 60})
	}
	if err != nil {
		return nil, "", err
	}

	return out.Bytes(), contentType, nil
}
