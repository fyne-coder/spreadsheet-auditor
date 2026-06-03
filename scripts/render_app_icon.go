package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const (
	outputSize = 1024
	scale      = 4
	canvasSize = outputSize * scale
)

func main() {
	out := filepath.Join("desktop", "build", "appicon.png")
	img := image.NewRGBA(image.Rect(0, 0, canvasSize, canvasSize))

	fillRoundedRect(img, 64, 64, 896, 896, 204, color.RGBA{11, 95, 255, 255}, color.RGBA{15, 118, 110, 255})
	fillCircle(img, 790, 234, 96, color.RGBA{34, 211, 238, 56})
	fillCircle(img, 210, 790, 120, color.RGBA{255, 255, 255, 28})

	fillRoundedRect(img, 226, 166, 552, 686, 64, color.RGBA{255, 255, 255, 255}, color.RGBA{232, 241, 255, 255})
	fillPolygon(img, []point{{642, 170}, {760, 288}, {682, 288}, {642, 248}}, color.RGBA{207, 228, 255, 255})
	fillRoundedRectSolid(img, 300, 360, 392, 308, 20, color.RGBA{248, 251, 255, 255})
	strokeRoundedRect(img, 300, 360, 392, 308, 20, 12, color.RGBA{199, 215, 254, 255})
	for _, y := range []float64{448, 532, 616} {
		strokeLine(img, 300, y, 692, y, 10, color.RGBA{188, 208, 234, 255})
	}
	for _, x := range []float64{398, 496, 594} {
		strokeLine(img, x, 360, x, 668, 10, color.RGBA{188, 208, 234, 255})
	}

	fillCircle(img, 650, 672, 132, color.RGBA{255, 255, 255, 240})
	strokeCircle(img, 650, 672, 92, 44, color.RGBA{15, 118, 110, 255})
	strokeLine(img, 724, 746, 816, 838, 54, color.RGBA{15, 118, 110, 255})
	strokeLine(img, 607, 673, 647, 713, 34, color.RGBA{16, 185, 129, 255})
	strokeLine(img, 647, 713, 725, 621, 34, color.RGBA{16, 185, 129, 255})

	fillCircle(img, 294, 300, 58, color.RGBA{245, 158, 11, 255})
	strokeLine(img, 294, 260, 294, 312, 20, color.RGBA{255, 255, 255, 255})
	fillCircle(img, 294, 338, 10, color.RGBA{255, 255, 255, 255})

	final := downsample(img)
	file, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := png.Encode(file, final); err != nil {
		panic(err)
	}
}

type point struct {
	x float64
	y float64
}

func fillRoundedRect(img *image.RGBA, x, y, w, h, r float64, top, bottom color.RGBA) {
	for py := int(y * scale); py < int((y+h)*scale); py++ {
		t := float64(py-int(y*scale)) / float64(int(h*scale))
		c := blendColor(top, bottom, t)
		for px := int(x * scale); px < int((x+w)*scale); px++ {
			if inRoundedRect(float64(px)/scale, float64(py)/scale, x, y, w, h, r) {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func fillRoundedRectSolid(img *image.RGBA, x, y, w, h, r float64, c color.RGBA) {
	fillRoundedRect(img, x, y, w, h, r, c, c)
}

func strokeRoundedRect(img *image.RGBA, x, y, w, h, r, width float64, c color.RGBA) {
	outer := width / 2
	inner := width / 2
	for py := int((y - outer) * scale); py < int((y+h+outer)*scale); py++ {
		for px := int((x - outer) * scale); px < int((x+w+outer)*scale); px++ {
			fx := float64(px) / scale
			fy := float64(py) / scale
			if inRoundedRect(fx, fy, x-outer, y-outer, w+width, h+width, r+outer) &&
				!inRoundedRect(fx, fy, x+inner, y+inner, w-width, h-width, math.Max(0, r-inner)) {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func inRoundedRect(px, py, x, y, w, h, r float64) bool {
	if px < x || px > x+w || py < y || py > y+h {
		return false
	}
	cx := math.Min(math.Max(px, x+r), x+w-r)
	cy := math.Min(math.Max(py, y+r), y+h-r)
	return math.Hypot(px-cx, py-cy) <= r
}

func fillCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	for py := int((cy - r) * scale); py < int((cy+r)*scale); py++ {
		for px := int((cx - r) * scale); px < int((cx+r)*scale); px++ {
			if math.Hypot(float64(px)/scale-cx, float64(py)/scale-cy) <= r {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func strokeCircle(img *image.RGBA, cx, cy, r, width float64, c color.RGBA) {
	for py := int((cy - r - width) * scale); py < int((cy+r+width)*scale); py++ {
		for px := int((cx - r - width) * scale); px < int((cx+r+width)*scale); px++ {
			d := math.Hypot(float64(px)/scale-cx, float64(py)/scale-cy)
			if d >= r-width/2 && d <= r+width/2 {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func strokeLine(img *image.RGBA, x1, y1, x2, y2, width float64, c color.RGBA) {
	minX := math.Min(x1, x2) - width
	maxX := math.Max(x1, x2) + width
	minY := math.Min(y1, y2) - width
	maxY := math.Max(y1, y2) + width
	dx := x2 - x1
	dy := y2 - y1
	lengthSq := dx*dx + dy*dy
	for py := int(minY * scale); py < int(maxY*scale); py++ {
		for px := int(minX * scale); px < int(maxX*scale); px++ {
			fx := float64(px) / scale
			fy := float64(py) / scale
			t := ((fx-x1)*dx + (fy-y1)*dy) / lengthSq
			t = math.Max(0, math.Min(1, t))
			projX := x1 + t*dx
			projY := y1 + t*dy
			if math.Hypot(fx-projX, fy-projY) <= width/2 {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func fillPolygon(img *image.RGBA, pts []point, c color.RGBA) {
	minX, maxX := pts[0].x, pts[0].x
	minY, maxY := pts[0].y, pts[0].y
	for _, p := range pts {
		minX = math.Min(minX, p.x)
		maxX = math.Max(maxX, p.x)
		minY = math.Min(minY, p.y)
		maxY = math.Max(maxY, p.y)
	}
	for py := int(minY * scale); py < int(maxY*scale); py++ {
		for px := int(minX * scale); px < int(maxX*scale); px++ {
			if inPolygon(float64(px)/scale, float64(py)/scale, pts) {
				blendPixel(img, px, py, c)
			}
		}
	}
}

func inPolygon(x, y float64, pts []point) bool {
	inside := false
	j := len(pts) - 1
	for i := range pts {
		if (pts[i].y > y) != (pts[j].y > y) &&
			x < (pts[j].x-pts[i].x)*(y-pts[i].y)/(pts[j].y-pts[i].y)+pts[i].x {
			inside = !inside
		}
		j = i
	}
	return inside
}

func blendPixel(img *image.RGBA, x, y int, src color.RGBA) {
	if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
		return
	}
	i := img.PixOffset(x, y)
	dst := color.RGBA{img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3]}
	a := float64(src.A) / 255
	outA := a + float64(dst.A)/255*(1-a)
	if outA == 0 {
		return
	}
	img.Pix[i] = uint8((float64(src.R)*a + float64(dst.R)*float64(dst.A)/255*(1-a)) / outA)
	img.Pix[i+1] = uint8((float64(src.G)*a + float64(dst.G)*float64(dst.A)/255*(1-a)) / outA)
	img.Pix[i+2] = uint8((float64(src.B)*a + float64(dst.B)*float64(dst.A)/255*(1-a)) / outA)
	img.Pix[i+3] = uint8(outA * 255)
}

func blendColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func downsample(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, outputSize, outputSize))
	for y := 0; y < outputSize; y++ {
		for x := 0; x < outputSize; x++ {
			var r, g, b, a int
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					i := src.PixOffset(x*scale+sx, y*scale+sy)
					r += int(src.Pix[i])
					g += int(src.Pix[i+1])
					b += int(src.Pix[i+2])
					a += int(src.Pix[i+3])
				}
			}
			n := scale * scale
			i := dst.PixOffset(x, y)
			dst.Pix[i] = uint8(r / n)
			dst.Pix[i+1] = uint8(g / n)
			dst.Pix[i+2] = uint8(b / n)
			dst.Pix[i+3] = uint8(a / n)
		}
	}
	return dst
}
