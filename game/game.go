package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	screenW    = 600
	screenH    = 660
	boardSize  = 360
	cellSize   = boardSize / 3 // 120
	boardX     = (screenW - boardSize) / 2
	boardY     = 170
	animFrames = 48 // ~0.8s at 60fps
)

var (
	colorBg   = color.RGBA{15, 15, 26, 255}
	colorGrid = color.RGBA{60, 60, 100, 255}
	colorX    = color.RGBA{255, 107, 107, 255}
	colorO    = color.RGBA{78, 205, 196, 255}
	colorWin  = color.RGBA{255, 217, 61, 255}
	colorDim  = color.RGBA{130, 130, 165, 255}
)

type phase int

const (
	phasePlaying phase = iota
	phaseAnim
)

type scores struct {
	xWins int
	oWins int
	draws int
}

func (s scores) total() int { return s.xWins + s.oWins + s.draws }

type Game struct {
	board      Board
	scores     scores
	phase      phase
	lastResult WinResult
	animTimer  int
	fontSrc    *text.GoTextFaceSource
}

func NewGame() *Game {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	return &Game{board: NewBoard(), fontSrc: src}
}

func (g *Game) face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: g.fontSrc, Size: size}
}

func (g *Game) Update() error {
	just := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	if g.phase == phaseAnim {
		g.animTimer++
		if g.animTimer >= animFrames || just {
			g.board.Reset()
			g.phase = phasePlaying
			g.animTimer = 0
		}
		return nil
	}

	if just && g.board.RandomMove() {
		if result, done := g.board.CheckResult(); done {
			switch {
			case result.isDraw:
				g.scores.draws++
			case result.winner == PlayerX:
				g.scores.xWins++
			default:
				g.scores.oWins++
			}
			g.lastResult = result
			g.phase = phaseAnim
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	g.drawHeader(screen)
	g.drawBoard(screen)
	g.drawStatus(screen)
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	drawTC(screen, "GAMES", g.face(15), screenW/2, 28, colorDim)
	drawTC(screen, fmt.Sprintf("%d", g.scores.total()), g.face(62), screenW/2, 90, colorWin)

	vector.StrokeLine(screen, 50, 123, float32(screenW-50), 123, 1, colorGrid, false)

	drawTC(screen, fmt.Sprintf("X  %d", g.scores.xWins), g.face(17), screenW/4, 148, colorX)
	drawTC(screen, fmt.Sprintf("DRAWS  %d", g.scores.draws), g.face(17), screenW/2, 148, colorDim)
	drawTC(screen, fmt.Sprintf("%d  O", g.scores.oWins), g.face(17), 3*screenW/4, 148, colorO)
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	vector.DrawFilledRect(screen,
		float32(boardX), float32(boardY), float32(boardSize), float32(boardSize),
		color.RGBA{22, 22, 38, 255}, false)

	lw := float32(2)
	for i := 1; i < 3; i++ {
		x := float32(boardX + i*cellSize)
		vector.StrokeLine(screen, x, float32(boardY), x, float32(boardY+boardSize), lw, colorGrid, false)
		y := float32(boardY + i*cellSize)
		vector.StrokeLine(screen, float32(boardX), y, float32(boardX+boardSize), y, lw, colorGrid, false)
	}
	vector.StrokeRect(screen,
		float32(boardX), float32(boardY), float32(boardSize), float32(boardSize),
		lw, colorGrid, false)

	for i, p := range g.board.cells {
		row, col := i/3, i%3
		cx := float32(boardX + col*cellSize + cellSize/2)
		cy := float32(boardY + row*cellSize + cellSize/2)
		switch p {
		case PlayerX:
			drawXPiece(screen, cx, cy, colorX)
		case PlayerO:
			drawOPiece(screen, cx, cy, colorO)
		}
	}

	// Flashing win line during animation
	if g.phase == phaseAnim && !g.lastResult.isDraw && (g.animTimer/5)%2 == 0 {
		line := g.lastResult.line
		r0, c0 := line[0]/3, line[0]%3
		r2, c2 := line[2]/3, line[2]%3
		x0 := float32(boardX + c0*cellSize + cellSize/2)
		y0 := float32(boardY + r0*cellSize + cellSize/2)
		x1 := float32(boardX + c2*cellSize + cellSize/2)
		y1 := float32(boardY + r2*cellSize + cellSize/2)
		vector.StrokeLine(screen, x0, y0, x1, y1, 8, colorWin, true)
	}
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	resultY := float64(boardY+boardSize) + 42
	hintY := float64(boardY+boardSize) + 72

	if g.phase == phaseAnim {
		var msg string
		var clr color.Color
		switch {
		case g.lastResult.isDraw:
			msg, clr = "DRAW", colorDim
		case g.lastResult.winner == PlayerX:
			msg, clr = "X WINS!", colorX
		default:
			msg, clr = "O WINS!", colorO
		}
		drawTC(screen, msg, g.face(32), screenW/2, resultY, clr)
		drawTC(screen, "SPACE or CLICK for new game", g.face(14), screenW/2, hintY, colorDim)
		return
	}
	drawTC(screen, "SPACE or CLICK to play", g.face(17), screenW/2, resultY+15, colorDim)
}

func (g *Game) Layout(_, _ int) (int, int) { return screenW, screenH }

func drawXPiece(screen *ebiten.Image, cx, cy float32, clr color.Color) {
	pad := float32(26)
	half := float32(cellSize / 2)
	vector.StrokeLine(screen, cx-half+pad, cy-half+pad, cx+half-pad, cy+half-pad, 6, clr, true)
	vector.StrokeLine(screen, cx+half-pad, cy-half+pad, cx-half+pad, cy+half-pad, 6, clr, true)
}

func drawOPiece(screen *ebiten.Image, cx, cy float32, clr color.Color) {
	vector.StrokeCircle(screen, cx, cy, float32(cellSize/2)-24, 6, clr, true)
}

func drawTC(screen *ebiten.Image, str string, face *text.GoTextFace, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, face, op)
}
