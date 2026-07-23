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
	switch v := img.(type) {
	case *image.NRGBA:
		return pixHasAlpha8(v.Pix)
	case *image.RGBA:
		return pixHasAlpha8(v.Pix)
	case *image.NRGBA64:
		return pixHasAlpha16(v.Pix)
	case *image.RGBA64:
		return pixHasAlpha16(v.Pix)
	default:
		return false
	}
}

func pixHasAlpha8(pix []byte) bool {
	for i := 3; i < len(pix); i += 4 {
		if pix[i] != 0xff {
			return true
		}
	}
	return false
}

func pixHasAlpha16(pix []byte) bool {
	for i := 6; i+1 < len(pix); i += 8 {
		if pix[i] != 0xff || pix[i+1] != 0xff {
			return true
		}
	}
	return false
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
