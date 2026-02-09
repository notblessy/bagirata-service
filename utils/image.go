package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	_ "image/png"

	"github.com/sirupsen/logrus"
)

const maxAvatarSize = 200
const jpegQuality = 80

// CompressImageBase64 decodes base64 image data, resizes to max maxAvatarSize, re-encodes as JPEG, returns base64.
// Supports JPEG and PNG input. Returns empty string on error.
func CompressImageBase64(base64Input string) string {
	if base64Input == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(base64Input)
	if err != nil {
		logrus.Warnf("avatar base64 decode: %v", err)
		return ""
	}
	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		logrus.Warnf("avatar image decode: %v", err)
		return ""
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return ""
	}
	// Scale down if needed
	if w > maxAvatarSize || h > maxAvatarSize {
		img = resizeImage(img, maxAvatarSize)
		if img == nil {
			return ""
		}
		bounds = img.Bounds()
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
		logrus.Warnf("avatar jpeg encode: %v", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// resizeImage scales img so the longest side is maxSize. Uses simple nearest-neighbor.
func resizeImage(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= maxSize && h <= maxSize {
		return img
	}
	var newW, newH int
	if w > h {
		newW = maxSize
		newH = h * maxSize / w
		if newH < 1 {
			newH = 1
		}
	} else {
		newH = maxSize
		newW = w * maxSize / h
		if newW < 1 {
			newW = 1
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * bounds.Dx() / newW
			srcY := y * bounds.Dy() / newH
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}
	return dst
}

