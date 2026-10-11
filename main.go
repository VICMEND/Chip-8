package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: chipGO <rom>")
		os.Exit(1)
	}

	chip := &Chip{}
	chip.loadSprites()
	if err := chip.loadROM(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "load rom:", err)
		os.Exit(1)
	}

	const scale = 10
	ebiten.SetWindowSize(64*scale, 32*scale)
	ebiten.SetWindowTitle("chipGO")

	if err := ebiten.RunGame(&Game{chip: chip}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
