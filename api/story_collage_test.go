package handler

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawImageCoverFillsTarget(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 20, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 20; x++ {
			src.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, 30, 30))
	target := image.Rect(5, 5, 25, 25)
	drawImageCover(dst, src, target)
	if got := dst.RGBAAt(15, 15); got.R != 255 || got.A != 255 {
		t.Fatalf("center pixel not filled by source image: %#v", got)
	}
	if got := dst.RGBAAt(1, 1); got.A != 0 {
		t.Fatalf("pixels outside target must remain untouched: %#v", got)
	}
}
