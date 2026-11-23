package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	_ "image/gif"
	"math"
)

// createSixPhotoStoryCollage builds one portrait 9:16 JPEG story from six Telegram photos.
func createSixPhotoStoryCollage(token string, fileIDs []string) ([]byte, error) {
	if len(fileIDs) != 6 {
		return nil, fmt.Errorf("يجب إرسال 6 صور بالضبط")
	}
	const canvasW, canvasH, gutter = 1080, 1920, 18
	cellW := (canvasW - 3*gutter) / 2
	cellH := (canvasH - 4*gutter) / 3
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	for i, fileID := range fileIDs {
		data, err := downloadTelegramFile(token, fileID)
		if err != nil {
			return nil, fmt.Errorf("تعذر تنزيل الصورة %d: %w", i+1, err)
		}
		src, format, err := image.Decode(bytes.NewReader(data))
		if err != nil || (format != "jpeg" && format != "png" && format != "gif") {
			return nil, fmt.Errorf("الصورة %d غير مدعومة أو تالفة", i+1)
		}
		col, row := i%2, i/2
		x0, y0 := gutter+col*(cellW+gutter), gutter+row*(cellH+gutter)
		drawImageCover(canvas, src, image.Rect(x0, y0, x0+cellW, y0+cellH))
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, canvas, &jpeg.Options{Quality: 92}); err != nil {
		return nil, fmt.Errorf("تعذر إنشاء صورة التجميعة: %w", err)
	}
	return out.Bytes(), nil
}

func drawImageCover(dst *image.RGBA, src image.Image, target image.Rectangle) {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	scale := math.Max(float64(target.Dx())/float64(sw), float64(target.Dy())/float64(sh))
	cropW := float64(target.Dx()) / scale
	cropH := float64(target.Dy()) / scale
	startX := float64(sb.Min.X) + (float64(sw)-cropW)/2
	startY := float64(sb.Min.Y) + (float64(sh)-cropH)/2
	for y := target.Min.Y; y < target.Max.Y; y++ {
		sy := int(startY + float64(y-target.Min.Y)/scale)
		for x := target.Min.X; x < target.Max.X; x++ {
			sx := int(startX + float64(x-target.Min.X)/scale)
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

// postBusinessStoryJPEG publishes the generated collage through the existing Business story connection.
func postBusinessStoryJPEG(token, businessConnID string, jpegData []byte, activePeriod string) error {
	if activePeriod == "" {
		activePeriod = "86400"
	}
	fields := map[string]string{
		"business_connection_id": businessConnID,
		"content":                `{"type":"photo","photo":"attach://content"}`,
		"active_period":          activePeriod,
		"post_to_chat_page":      "true",
	}
	return postMultipartBusinessAPI(token, "postStory", fields, "content", "story-collage.jpg", jpegData)
}

// Keep png referenced explicitly for decoder support across Go toolchains.
var _ = png.BestCompression
var _ color.Color
