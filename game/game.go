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
	screenH    = 720
	boardSize  = 360
	cellSize   = boardSize / 3 // 120
	boardX     = (screenW - boardSize) / 2
	boardY     = 170
	animFrames = 48 // ~0.8s at 60fps

	btnW      = 170
	btnH      = 76
	btnGap    = 15
	btnStartX = (screenW - (3*btnW + 2*btnGap)) / 2 // 30
	upgradeY  = 626
)

var (
	colorBg        = color.RGBA{15, 15, 26, 255}
	colorGrid      = color.RGBA{60, 60, 100, 255}
	colorX         = color.RGBA{255, 107, 107, 255}
	colorO         = color.RGBA{78, 205, 196, 255}
	colorWin       = color.RGBA{255, 217, 61, 255}
	colorDim       = color.RGBA{130, 130, 165, 255}
	colorBtnBg     = color.RGBA{25, 25, 45, 255}
	colorBtnBorder = color.RGBA{70, 70, 115, 255}
	colorAfford    = color.RGBA{100, 220, 110, 255}
	colorMaxed     = color.RGBA{160, 100, 240, 255}
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
	board        Board
	scores       scores
	phase        phase
	lastResult   WinResult
	animTimer    int
	fontSrc      *text.GoTextFaceSource
	currency     int
	moreTicLevel int
	idleTimer    int
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

// moreTicCost returns the cost for the next more-tic level (starts at 1, +3 per level).
func moreTicCost(level int) int { return 1 + level*3 }

func (g *Game) recordResult(result WinResult) {
	switch {
	case result.isDraw:
		g.scores.draws++
	case result.winner == PlayerX:
		g.scores.xWins++
	default:
		g.scores.oWins++
	}
	g.currency++
	g.lastResult = result
	g.phase = phaseAnim
	g.idleTimer = 0
}

// tryBuyUpgrade checks if a mouse click landed on an upgrade button and handles it.
// Returns true if the click was consumed by a button area.
func (g *Game) tryBuyUpgrade(mx, my int) bool {
	if my < upgradeY || my >= upgradeY+btnH {
		return false
	}
	bx0 := btnStartX
	bx1 := btnStartX + btnW + btnGap
	bx2 := btnStartX + 2*(btnW+btnGap)

	switch {
	case mx >= bx0 && mx < bx0+btnW:
		if g.moreTicLevel < 3 {
			cost := moreTicCost(g.moreTicLevel)
			if g.currency >= cost {
				g.currency -= cost
				g.moreTicLevel++
				g.idleTimer = 0
			}
		}
		return true
	case mx >= bx1 && mx < bx1+btnW:
		return true // more tac placeholder
	case mx >= bx2 && mx < bx2+btnW:
		return true // more toe placeholder
	}
	return false
}

func (g *Game) Update() error {
	mouseClicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if mouseClicked {
		mx, my := ebiten.CursorPosition()
		if g.tryBuyUpgrade(mx, my) {
			mouseClicked = false
		}
	}

	just := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		mouseClicked

	if g.phase == phaseAnim {
		g.animTimer++
		if g.animTimer >= animFrames || just {
			g.board.Reset()
			g.phase = phasePlaying
			g.animTimer = 0
		}
		return nil
	}

	// Idle auto-moves at moreTicLevel moves/sec across all boards (currently one).
	if g.moreTicLevel > 0 {
		g.idleTimer++
		if g.idleTimer >= 60/g.moreTicLevel {
			g.idleTimer = 0
			if g.board.RandomMove() {
				if result, done := g.board.CheckResult(); done {
					g.recordResult(result)
					return nil
				}
			}
		}
	}

	if just {
		if g.board.RandomMove() {
			if result, done := g.board.CheckResult(); done {
				g.recordResult(result)
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	g.drawHeader(screen)
	g.drawBoard(screen)
	g.drawStatus(screen)
	g.drawUpgrades(screen)
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	drawTC(screen, "GAMES", g.face(15), screenW/2, 28, colorDim)
	drawTC(screen, fmt.Sprintf("%d", g.currency), g.face(62), screenW/2, 90, colorWin)

	vector.StrokeLine(screen, 50, 123, float32(screenW-50), 123, 1, colorGrid, false)

	drawTC(screen, fmt.Sprintf("X  %d", g.scores.xWins), g.face(17), screenW/4, 148, colorX)
	drawTC(screen, fmt.Sprintf("DRAWS  %d", g.scores.draws), g.face(17), screenW/2, 148, colorDim)
	drawTC(screen, fmt.Sprintf("%d  O", g.scores.oWins), g.face(17), 3*screenW/4, 148, colorO)
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	// TODO: cleanup deprecated function references
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

	// Might be a gamedev thing but also feels like optimization problem
	// Paint entire board state to then display at once but may be splitting hairs
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

	// Solid win line during animation, in the winner's color
	if g.phase == phaseAnim && !g.lastResult.isDraw {
		line := g.lastResult.line
		r0, c0 := line[0]/3, line[0]%3
		r2, c2 := line[2]/3, line[2]%3
		x0 := float32(boardX + c0*cellSize + cellSize/2)
		y0 := float32(boardY + r0*cellSize + cellSize/2)
		x1 := float32(boardX + c2*cellSize + cellSize/2)
		y1 := float32(boardY + r2*cellSize + cellSize/2)
		winColor := colorX
		if g.lastResult.winner == PlayerO {
			winColor = colorO
		}
		vector.StrokeLine(screen, x0, y0, x1, y1, 8, winColor, true)
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

func (g *Game) drawUpgrades(screen *ebiten.Image) {
	drawTC(screen, "UPGRADES", g.face(13), screenW/2, float64(upgradeY)-14, colorDim)
	g.drawMoreTicBtn(screen)
	g.drawLockedBtn(screen, btnStartX+btnW+btnGap, "MORE TAC")
	g.drawLockedBtn(screen, btnStartX+2*(btnW+btnGap), "MORE TOE")
}

func (g *Game) drawMoreTicBtn(screen *ebiten.Image) {
	x, y := float32(btnStartX), float32(upgradeY)
	cx := float64(btnStartX + btnW/2)
	cy := float64(upgradeY)

	canAfford := g.currency >= moreTicCost(g.moreTicLevel)

	bgClr := color.Color(colorBtnBg)
	bgClr = color.RGBA{20, 40, 25, 255}
	borderClr := color.Color(colorBtnBorder)
	borderClr = colorAfford

	vector.DrawFilledRect(screen, x, y, float32(btnW), float32(btnH), bgClr, false)
	vector.StrokeRect(screen, x, y, float32(btnW), float32(btnH), 1.5, borderClr, false)

	drawTC(screen, "MORE TIC", g.face(14), cx, cy+16, colorWin)

	cost := moreTicCost(g.moreTicLevel)
	costLabel := fmt.Sprintf("COST: %d game", cost)
	if cost != 1 {
		costLabel += "s"
	}
	costClr := color.Color(colorDim)
	if canAfford {
		costClr = colorAfford
	}
	drawTC(screen, costLabel, g.face(11), cx, cy+36, costClr)
	if g.moreTicLevel == 0 {
		drawTC(screen, "unlock auto-move", g.face(11), cx, cy+56, colorDim)
	} else {
		drawTC(screen, fmt.Sprintf("LVL %d · %d move/sec", g.moreTicLevel, g.moreTicLevel), g.face(11), cx, cy+56, colorDim)
	}
}

func (g *Game) drawLockedBtn(screen *ebiten.Image, bx int, label string) {
	x, y := float32(bx), float32(upgradeY)
	cx := float64(bx + btnW/2)
	cy := float64(upgradeY)

	vector.DrawFilledRect(screen, x, y, float32(btnW), float32(btnH), colorBtnBg, false)
	vector.StrokeRect(screen, x, y, float32(btnW), float32(btnH), 1.5, colorBtnBorder, false)

	drawTC(screen, label, g.face(14), cx, cy+28, colorDim)
	drawTC(screen, "LOCKED", g.face(11), cx, cy+50, color.RGBA{70, 70, 100, 255})
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
