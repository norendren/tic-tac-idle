package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Tic-Tac-Idle")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
