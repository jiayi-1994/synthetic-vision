package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"

	_ "image/jpeg"
	_ "golang.org/x/image/webp"
)

// compositeMaskedEdit overlays the provider edit result onto the original
// source using OpenAI-style mask semantics: transparent mask pixels are editable
// and opaque pixels are preserved from the source. This guarantees that local
// editing never rewrites areas the user did not paint, even if the provider
// returns a full-frame image.
func compositeMaskedEdit(resultBytes, sourceBytes, maskBytes []byte) ([]byte, error) {
	resultImg, _, err := image.Decode(bytes.NewReader(resultBytes))
	if err != nil {
		return nil, err
	}
	sourceImg, _, err := image.Decode(bytes.NewReader(sourceBytes))
	if err != nil {
		return nil, err
	}
	maskImg, _, err := image.Decode(bytes.NewReader(maskBytes))
	if err != nil {
		return nil, err
	}

	sourceBounds := sourceImg.Bounds()
	w, h := sourceBounds.Dx(), sourceBounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, errors.New("source image has empty bounds")
	}

	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			src := sampleNRGBA(sourceImg, x, y, w, h)
			edited := sampleNRGBA(resultImg, x, y, w, h)
			maskAlpha := sampleAlpha(maskImg, x, y, w, h)

			editable := 1.0 - float64(maskAlpha)/255.0
			out.SetNRGBA(x, y, blendNRGBA(src, edited, editable))
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sampleNRGBA(img image.Image, dstX, dstY, dstW, dstH int) color.NRGBA {
	b := img.Bounds()
	x := b.Min.X + scaledCoord(dstX, dstW, b.Dx())
	y := b.Min.Y + scaledCoord(dstY, dstH, b.Dy())
	return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
}

func sampleAlpha(img image.Image, dstX, dstY, dstW, dstH int) uint8 {
	b := img.Bounds()
	x := b.Min.X + scaledCoord(dstX, dstW, b.Dx())
	y := b.Min.Y + scaledCoord(dstY, dstH, b.Dy())
	_, _, _, a := img.At(x, y).RGBA()
	return uint8(a >> 8)
}

func scaledCoord(dstCoord, dstSize, srcSize int) int {
	if srcSize <= 1 || dstSize <= 1 {
		return 0
	}
	v := dstCoord * srcSize / dstSize
	if v >= srcSize {
		return srcSize - 1
	}
	if v < 0 {
		return 0
	}
	return v
}

func blendNRGBA(source, edited color.NRGBA, editable float64) color.NRGBA {
	if editable <= 0 {
		return source
	}
	if editable >= 1 {
		return edited
	}
	preserve := 1.0 - editable
	return color.NRGBA{
		R: uint8(float64(source.R)*preserve + float64(edited.R)*editable + 0.5),
		G: uint8(float64(source.G)*preserve + float64(edited.G)*editable + 0.5),
		B: uint8(float64(source.B)*preserve + float64(edited.B)*editable + 0.5),
		A: uint8(float64(source.A)*preserve + float64(edited.A)*editable + 0.5),
	}
}
