package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	cyclesPerFrame = 12
)

var keyMap = map[ebiten.Key]uint8{
	ebiten.Key1: 0x1,
	ebiten.Key2: 0x2,
	ebiten.Key3: 0x3,
	ebiten.Key4: 0xC,

	ebiten.KeyQ: 0x4,
	ebiten.KeyW: 0x5,
	ebiten.KeyE: 0x6,
	ebiten.KeyR: 0xD,

	ebiten.KeyA: 0x7,
	ebiten.KeyS: 0x8,
	ebiten.KeyD: 0x9,
	ebiten.KeyF: 0xE,

	ebiten.KeyZ: 0xA,
	ebiten.KeyX: 0x0,
	ebiten.KeyC: 0xB,
	ebiten.KeyV: 0xF,
}

type Game struct {
	chip *Chip
}

func (g *Game) Update() error {
	for i := range g.chip.keypad {
		g.chip.keypad[i] = 0
	}

	for k, v := range keyMap {
		if ebiten.IsKeyPressed(k) {
			g.chip.keypad[v] = 1
		}
	}

	for range cyclesPerFrame {
		g.chip.cycle()
		if g.chip.opcode&0xF000 == 0xD000 {
			break
		}
	}
	g.chip.tickTimers()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	for y := range 32 {
		for x := range 64 {
			if g.chip.display[y*64+x] == 1 {
				screen.Set(x, y, color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 64, 32
}
