package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestCompositeMaskedEditPreservesUnselectedPixels(t *testing.T) {
	source := encodeSolidPNG(t, 4, 4, color.NRGBA{R: 240, G: 20, B: 20, A: 255})
	edited := encodeSolidPNG(t, 4, 4, color.NRGBA{R: 20, G: 60, B: 240, A: 255})
	mask := encodeMaskPNG(t, 4, 4, func(x, y int) uint8 {
		if x >= 1 && x <= 2 && y >= 1 && y <= 2 {
			return 0
		}
		return 255
	})

	outBytes, err := compositeMaskedEdit(edited, source, mask)
	if err != nil {
		t.Fatalf("compositeMaskedEdit returned error: %v", err)
	}
	out := decodePNG(t, outBytes)

	assertNRGBAAt(t, out, 0, 0, color.NRGBA{R: 240, G: 20, B: 20, A: 255})
	assertNRGBAAt(t, out, 3, 3, color.NRGBA{R: 240, G: 20, B: 20, A: 255})
	assertNRGBAAt(t, out, 1, 1, color.NRGBA{R: 20, G: 60, B: 240, A: 255})
	assertNRGBAAt(t, out, 2, 2, color.NRGBA{R: 20, G: 60, B: 240, A: 255})
}

func TestCompositeMaskedEditScalesProviderResultToSourceSize(t *testing.T) {
	source := encodeSolidPNG(t, 4, 4, color.NRGBA{R: 200, G: 10, B: 10, A: 255})
	edited := encodeSolidPNG(t, 2, 2, color.NRGBA{R: 10, G: 90, B: 210, A: 255})
	mask := encodeMaskPNG(t, 2, 2, func(x, y int) uint8 { return 0 })

	outBytes, err := compositeMaskedEdit(edited, source, mask)
	if err != nil {
		t.Fatalf("compositeMaskedEdit returned error: %v", err)
	}
	out := decodePNG(t, outBytes)

	if got := out.Bounds().Dx(); got != 4 {
		t.Fatalf("width = %d, want 4", got)
	}
	if got := out.Bounds().Dy(); got != 4 {
		t.Fatalf("height = %d, want 4", got)
	}
	assertNRGBAAt(t, out, 3, 3, color.NRGBA{R: 10, G: 90, B: 210, A: 255})
}

func encodeSolidPNG(t *testing.T, w, h int, c color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return encodePNG(t, img)
}

func encodeMaskPNG(t *testing.T, w, h int, alpha func(x, y int) uint8) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, color.NRGBA{A: alpha(x, y)})
		}
	}
	return encodePNG(t, img)
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func decodePNG(t *testing.T, b []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	return img
}

func assertNRGBAAt(t *testing.T, img image.Image, x, y int, want color.NRGBA) {
	t.Helper()
	got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	if got != want {
		t.Fatalf("pixel (%d,%d) = %#v, want %#v", x, y, got, want)
	}
}
