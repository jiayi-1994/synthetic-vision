package provider

import (
	"bytes"
	"context"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"time"

	"syntheticvision/internal/config"
)

// MockProvider renders a deterministic, on-brand abstract image purely from the
// stdlib. The same Seed + Prompt always produces the same PNG.
type MockProvider struct {
	delayMs int
}

// NewMockProvider constructs a MockProvider honoring cfg.MockDelayMs.
func NewMockProvider(cfg config.Config) *MockProvider {
	return &MockProvider{delayMs: cfg.MockDelayMs}
}

// Name implements Provider.
func (m *MockProvider) Name() string { return "mock" }

// darkPalette holds the deep base tones used as the gradient origin so every
// render has depth (no two-darks-blending-flat results).
var darkPalette = []color.RGBA{
	{0x06, 0x0e, 0x20, 0xff}, // #060e20 surface-container-lowest
	{0x0b, 0x13, 0x26, 0xff}, // #0b1326 background
	{0x13, 0x1b, 0x2e, 0xff}, // #131b2e surface-container-low
	{0x17, 0x1f, 0x33, 0xff}, // #171f33 surface-container
	{0x3f, 0x00, 0x8e, 0xff}, // #3f008e on-primary (deep violet)
	{0x00, 0x34, 0x4e, 0xff}, // #00344e deep cyan
}

// vividPalette holds the bright accents that the gradient resolves toward and
// that the radial glow blobs are drawn from.
var vividPalette = []color.RGBA{
	{0x7c, 0x3a, 0xed, 0xff}, // #7c3aed primary-container
	{0xd2, 0xbb, 0xff, 0xff}, // #d2bbff primary
	{0x00, 0xa2, 0xe6, 0xff}, // #00a2e6 secondary-container
	{0x89, 0xce, 0xff, 0xff}, // #89ceff secondary
	{0xa1, 0x51, 0x00, 0xff}, // #a15100 tertiary-container
	{0xff, 0xb7, 0x84, 0xff}, // #ffb784 tertiary
	{0x73, 0x2e, 0xe4, 0xff}, // #732ee4 inverse-primary
}

// glowColors are bright accents blended additively for the radial blobs.
var glowColors = vividPalette

// vignetteColor is the darkest surface tone the edges fade toward.
var vignetteColor = color.RGBA{0x06, 0x0e, 0x20, 0xff} // #060e20

// Generate produces the PNG. It splits the configured delay into short slices
// so a cancelled context aborts promptly.
func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if err := m.sleep(ctx); err != nil {
		return nil, err
	}

	h := fnv.New64a()
	_, _ = h.Write([]byte(req.Prompt))
	_, _ = h.Write([]byte(req.Mode))
	if req.SourceImage != nil {
		_, _ = h.Write(req.SourceImage.Data[:min(len(req.SourceImage.Data), 4096)])
	}
	if req.MaskImage != nil {
		_, _ = h.Write(req.MaskImage.Data[:min(len(req.MaskImage.Data), 4096)])
	}
	seed := req.Seed ^ int64(h.Sum64())
	rng := rand.New(rand.NewSource(seed))

	w, h2 := req.Width, req.Height
	if w <= 0 {
		w = 1024
	}
	if h2 <= 0 {
		h2 = 1024
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h2))

	// 1. Base gradient: a deep base tone resolving toward a vivid accent so
	//    every render has both depth and color.
	cA := darkPalette[rng.Intn(len(darkPalette))]
	cB := vividPalette[rng.Intn(len(vividPalette))]
	// gradient direction: 0 vertical, 1 horizontal, 2 diagonal.
	dir := rng.Intn(3)
	drawGradient(img, cA, cB, dir)

	// 2. Radial glow blobs (4–7), additive toward an accent, plus one brighter
	//    focal "energy core" so the composition has a subject.
	blobs := 4 + rng.Intn(4)
	type blob struct {
		cx, cy, r float64
		col       color.RGBA
		strength  float64
	}
	bs := make([]blob, blobs)
	maxDim := float64(max(w, h2))
	for i := range bs {
		bs[i] = blob{
			cx:       rng.Float64() * float64(w),
			cy:       rng.Float64() * float64(h2),
			r:        maxDim * (0.20 + rng.Float64()*0.45),
			col:      glowColors[rng.Intn(len(glowColors))],
			strength: 0.60 + rng.Float64()*0.65,
		}
	}
	// Focal core: a tighter, brighter blob placed off-center (rule-of-thirds-ish).
	bs[0] = blob{
		cx:       float64(w) * (0.30 + rng.Float64()*0.40),
		cy:       float64(h2) * (0.30 + rng.Float64()*0.40),
		r:        maxDim * (0.12 + rng.Float64()*0.16),
		col:      vividPalette[rng.Intn(3)], // primary / primary-bright / secondary-container
		strength: 1.10 + rng.Float64()*0.5,
	}
	// Apply blobs with additive-ish blend.
	for y := 0; y < h2; y++ {
		for x := 0; x < w; x++ {
			base := img.RGBAAt(x, y)
			rf := float64(base.R)
			gf := float64(base.G)
			bf := float64(base.B)
			for _, bl := range bs {
				dx := float64(x) - bl.cx
				dy := float64(y) - bl.cy
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist >= bl.r {
					continue
				}
				// smooth falloff (cosine-ish)
				t := 1.0 - dist/bl.r
				f := t * t * bl.strength
				rf += float64(bl.col.R) * f
				gf += float64(bl.col.G) * f
				bf += float64(bl.col.B) * f
			}
			img.SetRGBA(x, y, color.RGBA{
				R: clamp8(rf),
				G: clamp8(gf),
				B: clamp8(bf),
				A: 0xff,
			})
		}
	}

	// 3. Faint geometric overlay: concentric rings or thin lines.
	if rng.Intn(2) == 0 {
		drawRings(img, rng, w, h2)
	} else {
		drawLines(img, rng, w, h2)
	}
	if req.SourceImage != nil {
		drawReferenceEditOverlay(img, w, h2, req.MaskImage != nil)
	}

	// 4. Subtle per-pixel noise.
	addNoise(img, rng, w, h2)

	// 5. Vignette darkening toward the edges.
	applyVignette(img, w, h2)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return &GenerateResult{Image: buf.Bytes(), MimeType: "image/png"}, nil
}

func (m *MockProvider) sleep(ctx context.Context) error {
	remaining := time.Duration(m.delayMs) * time.Millisecond
	if remaining <= 0 {
		return ctx.Err()
	}
	const slice = 50 * time.Millisecond
	for remaining > 0 {
		step := slice
		if step > remaining {
			step = remaining
		}
		t := time.NewTimer(step)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
		remaining -= step
	}
	return nil
}

func drawGradient(img *image.RGBA, a, b color.RGBA, dir int) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	denomV := float64(h - 1)
	denomH := float64(w - 1)
	denomD := float64(w + h - 2)
	if denomV <= 0 {
		denomV = 1
	}
	if denomH <= 0 {
		denomH = 1
	}
	if denomD <= 0 {
		denomD = 1
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var t float64
			switch dir {
			case 0: // vertical
				t = float64(y) / denomV
			case 1: // horizontal
				t = float64(x) / denomH
			default: // diagonal
				t = (float64(x) + float64(y)) / denomD
			}
			img.SetRGBA(x, y, lerpColor(a, b, t))
		}
	}
}

func drawRings(img *image.RGBA, rng *rand.Rand, w, h int) {
	cx := rng.Float64() * float64(w)
	cy := rng.Float64() * float64(h)
	rings := 4 + rng.Intn(5)
	maxR := float64(max(w, h)) * 0.55
	col := glowColors[rng.Intn(len(glowColors))]
	for i := 1; i <= rings; i++ {
		radius := maxR * float64(i) / float64(rings)
		drawCircleOutline(img, cx, cy, radius, col, 0.10)
	}
}

func drawCircleOutline(img *image.RGBA, cx, cy, radius float64, col color.RGBA, alpha float64) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	thickness := 1.5
	x0 := int(math.Floor(cx - radius - thickness))
	x1 := int(math.Ceil(cx + radius + thickness))
	y0 := int(math.Floor(cy - radius - thickness))
	y1 := int(math.Ceil(cy + radius + thickness))
	for y := y0; y <= y1; y++ {
		if y < 0 || y >= h {
			continue
		}
		for x := x0; x <= x1; x++ {
			if x < 0 || x >= w {
				continue
			}
			d := math.Abs(math.Hypot(float64(x)-cx, float64(y)-cy) - radius)
			if d <= thickness {
				blendOver(img, x, y, col, alpha*(1.0-d/thickness))
			}
		}
	}
}

func drawLines(img *image.RGBA, rng *rand.Rand, w, h int) {
	lines := 3 + rng.Intn(4)
	col := glowColors[rng.Intn(len(glowColors))]
	for i := 0; i < lines; i++ {
		x0 := rng.Float64() * float64(w)
		y0 := rng.Float64() * float64(h)
		ang := rng.Float64() * 2 * math.Pi
		length := float64(max(w, h))
		x1 := x0 + math.Cos(ang)*length
		y1 := y0 + math.Sin(ang)*length
		drawLine(img, x0, y0, x1, y1, col, 0.08)
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 float64, col color.RGBA, alpha float64) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	steps := int(math.Hypot(x1-x0, y1-y0))
	if steps <= 0 {
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(x0 + (x1-x0)*t))
		y := int(math.Round(y0 + (y1-y0)*t))
		if x < 0 || x >= w || y < 0 || y >= h {
			continue
		}
		blendOver(img, x, y, col, alpha)
	}
}

func drawReferenceEditOverlay(img *image.RGBA, w, h int, masked bool) {
	cyan := color.RGBA{0x38, 0xe8, 0xff, 0xff}
	magenta := color.RGBA{0xff, 0x3d, 0xf0, 0xff}
	margin := max(10, min(w, h)/24)
	for i := 0; i < 3; i++ {
		alpha := 0.28 - float64(i)*0.06
		x0 := margin + i*5
		y0 := margin + i*5
		x1 := w - margin - i*5
		y1 := h - margin - i*5
		for x := x0; x < x1; x++ {
			blendOver(img, x, y0, cyan, alpha)
			blendOver(img, x, y1, cyan, alpha)
		}
		for y := y0; y < y1; y++ {
			blendOver(img, x0, y, cyan, alpha)
			blendOver(img, x1, y, cyan, alpha)
		}
	}
	if !masked {
		return
	}
	step := max(18, min(w, h)/18)
	for x := -h; x < w+h; x += step {
		drawLine(img, float64(x), 0, float64(x+h), float64(h), magenta, 0.13)
	}
}

func addNoise(img *image.RGBA, rng *rand.Rand, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			n := float64(rng.Intn(13) - 6) // ±6
			c := img.RGBAAt(x, y)
			img.SetRGBA(x, y, color.RGBA{
				R: clamp8(float64(c.R) + n),
				G: clamp8(float64(c.G) + n),
				B: clamp8(float64(c.B) + n),
				A: 0xff,
			})
		}
	}
}

func applyVignette(img *image.RGBA, w, h int) {
	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	maxDist := math.Hypot(cx, cy)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxDist
			// only darken in the outer region.
			v := (d - 0.55) / 0.45
			if v <= 0 {
				continue
			}
			if v > 1 {
				v = 1
			}
			f := v * v * 0.75 // max 75% toward vignette color
			c := img.RGBAAt(x, y)
			img.SetRGBA(x, y, lerpColor(c, vignetteColor, f))
		}
	}
}

func blendOver(img *image.RGBA, x, y int, col color.RGBA, alpha float64) {
	if alpha <= 0 {
		return
	}
	if alpha > 1 {
		alpha = 1
	}
	base := img.RGBAAt(x, y)
	img.SetRGBA(x, y, lerpColor(base, col, alpha))
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 0xff,
	}
}

func clamp8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
