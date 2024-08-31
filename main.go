package main

import (
	_ "embed"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	WIDTH  = 1024
	HEIGHT = 768
)

//go:embed shader.kage
var shaderProgram []byte

func main() {
	// create shader object
	shader, err := ebiten.NewShader(shaderProgram)
	if err != nil {
		log.Fatalf("Failed to compile shader.kage: %v", err)
	}

	// create game struct
	game := &Game{
		shader: shader,
		active: ebiten.NewImage(WIDTH, HEIGHT),
		buff:   ebiten.NewImage(WIDTH, HEIGHT),
		scale:  1,
	}

	ebiten.SetWindowTitle("intro/pixelize")
	ebiten.SetWindowSize(WIDTH, HEIGHT)
	ebiten.SetTPS(100)
	for y := 10; y < 20; y++ {
		for x := 100; x < 130; x++ {
			game.active.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	err = ebiten.RunGame(game)
	if err != nil {
		log.Fatal(err)
	}
}

// Struct implementing the ebiten.Game interface.
type Game struct {
	shader   *ebiten.Shader
	vertices [4]ebiten.Vertex
	active   *ebiten.Image
	buff     *ebiten.Image
	mousePX  int
	mousePY  int

	scale   float32
	offsetx int
	offsety int

	frame uint64
	pause bool
}

// Assume a fixed layout.
func (self *Game) Layout(_, _ int) (int, int) {
	return WIDTH, HEIGHT
}

// No logic to update.

func (self *Game) Update() error {
	self.frame++
	mx, my := ebiten.CursorPosition()
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		self.pause = !self.pause
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		self.active.Clear()
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		bounds := self.active.Bounds()
		scaledWidth := float32(bounds.Dx()) / self.scale
		scaledHeight := float32(bounds.Dy()) / self.scale

		mmx := self.offsetx + int(scaledWidth*float32(mx)/WIDTH)
		mmy := self.offsety + int(scaledHeight*float32(my)/HEIGHT)

		ox := self.offsetx + int(scaledWidth*float32(self.mousePX)/WIDTH)
		oy := self.offsety + int(scaledHeight*float32(self.mousePY)/HEIGHT)

		doLine(mmx, mmy, ox, oy, func(cx, cy int) {
			self.active.Set(cx, cy, color.RGBA{255, 0, 0, 255})
		})
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		bounds := self.active.Bounds()
		scaledWidth := float32(bounds.Dx()) / self.scale
		scaledHeight := float32(bounds.Dy()) / self.scale

		mmx := self.offsetx + int(scaledWidth*float32(mx)/WIDTH)
		mmy := self.offsety + int(scaledHeight*float32(my)/HEIGHT)

		ox := self.offsetx + int(scaledWidth*float32(self.mousePX)/WIDTH)
		oy := self.offsety + int(scaledHeight*float32(self.mousePY)/HEIGHT)

		doLine(mmx, mmy, ox, oy, func(cx, cy int) {
			self.active.Set(cx, cy, color.RGBA{0, 0, 0, 0})
		})
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {

		self.offsetx -= (mx - self.mousePX) / 2
		self.offsety -= (my - self.mousePY) / 2
		if self.offsetx < 0 {
			self.offsetx = 0
		}
		if self.offsety < 0 {
			self.offsety = 0
		}

		if self.offsetx+int(float32(WIDTH)/self.scale) > WIDTH {
			self.offsetx -= self.offsetx + int(float32(WIDTH)/self.scale) - WIDTH
		}
		if self.offsety+int(float32(HEIGHT)/self.scale) > HEIGHT {
			self.offsety -= self.offsety + int(float32(HEIGHT)/self.scale) - HEIGHT
		}
	}
	if _, y := ebiten.Wheel(); y != 0 {
		self.scale += float32(y) / 10
		if self.scale < 1 {
			self.scale = 1
		}
		if self.scale > 10 {
			self.scale = 10
		}
	}
	self.mousePX = mx
	self.mousePY = my

	if self.frame%10 == 0 && !self.pause {
		//if !self.pause {
		self.buff.Clear()
		// map the vertices to the target image
		bounds := self.buff.Bounds()
		self.vertices[0].DstX = float32(bounds.Min.X) // top-left
		self.vertices[0].DstY = float32(bounds.Min.Y) // top-left
		self.vertices[1].DstX = float32(bounds.Max.X) // top-right
		self.vertices[1].DstY = float32(bounds.Min.Y) // top-right
		self.vertices[2].DstX = float32(bounds.Min.X) // bottom-left
		self.vertices[2].DstY = float32(bounds.Max.Y) // bottom-left
		self.vertices[3].DstX = float32(bounds.Max.X) // bottom-right
		self.vertices[3].DstY = float32(bounds.Max.Y) // bottom-right

		// set the source image sampling coordinates
		srcBounds := self.active.Bounds()
		self.vertices[0].SrcX = float32(srcBounds.Min.X) // top-left
		self.vertices[0].SrcY = float32(srcBounds.Min.Y) // top-left
		self.vertices[1].SrcX = float32(srcBounds.Max.X) // top-right
		self.vertices[1].SrcY = float32(srcBounds.Min.Y) // top-right
		self.vertices[2].SrcX = float32(srcBounds.Min.X) // bottom-left
		self.vertices[2].SrcY = float32(srcBounds.Max.Y) // bottom-left
		self.vertices[3].SrcX = float32(srcBounds.Max.X) // bottom-right
		self.vertices[3].SrcY = float32(srcBounds.Max.Y) // bottom-right

		// triangle shader options
		var shaderOpts ebiten.DrawTrianglesShaderOptions
		shaderOpts.Images[0] = self.active

		// draw shader
		indices := []uint16{0, 1, 2, 2, 1, 3} // map vertices to triangles
		self.buff.DrawTrianglesShader(self.vertices[:], indices, self.shader, &shaderOpts)
		self.active, self.buff = self.buff, self.active
	}
	return nil

}

// Core drawing function from where we call DrawTrianglesShader.
func (self *Game) Draw(screen *ebiten.Image) {

	var dio ebiten.DrawImageOptions
	dio.GeoM.Scale(float64(self.scale), float64(self.scale))
	bounds := self.active.Bounds()
	screen.DrawImage(
		self.active.SubImage(
			image.Rect(
				self.offsetx,
				self.offsety,
				self.offsetx+int(float32(bounds.Dx())/self.scale),
				self.offsety+int(float32(bounds.Dy())/self.scale),
			)).(*ebiten.Image), &dio)
}

func absi(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// Bresenham's line algorithm
func doLine(x0, y0, x1, y1 int, f func(x, y int)) {
	dx := absi(x1 - x0)
	var sx int
	if x0 < x1 {
		sx = 1
	} else {
		sx = -1
	}
	dy := -absi(y1 - y0)
	var sy int
	if y0 < y1 {
		sy = 1
	} else {
		sy = -1
	}
	error := dx + dy

	for {
		f(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * error
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			error = error + dy
			x0 = x0 + sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			error = error + dx
			y0 = y0 + sy
		}
	}
}
