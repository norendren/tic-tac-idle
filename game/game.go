package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

const (
	screenW = 600
	screenH = 720

	// Board area: fixed region boards render into, regardless of count.
	boardAreaX = 10
	boardAreaY = 162
	boardAreaW = screenW - 2*boardAreaX // 580
	boardAreaH = 368                    // 530 - boardAreaY

	animFrames = 48  // ~0.8s at 60fps
	resultTTL  = 120 // frames to show result label in status bar

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
)

// boardSlot holds the state for a single board instance.
type boardSlot struct {
	board     Board
	animTimer int // >0 means showing result animation
	result    WinResult
}

func (s *boardSlot) inAnim() bool { return s.animTimer > 0 }

type scores struct {
	xWins int
	oWins int
	draws int
}

func (s scores) total() int { return s.xWins + s.oWins + s.draws }

type Game struct {
	slots        []boardSlot
	scores       scores
	currency     int
	moreTicLevel int // auto-move rate: moreTicLevel moves/sec across all boards
	moreTacLevel int // number of extra boards purchased
	idleAccum    int // accumulator for fractional/multi moves per frame
	recentResult WinResult
	recentTimer  int
	fontSrc      *text.GoTextFaceSource
}

func NewGame() *Game {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatal(err)
	}
	return &Game{
		slots:   []boardSlot{{board: NewBoard()}},
		fontSrc: src,
	}
}

func (g *Game) face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: g.fontSrc, Size: size}
}

// func moreTicCost(level int) int { return 1 + level*3 }
func moreTicCost(level int) int { return 0 }
func moreTacCost(level int) int { return 0 }

//func moreTacCost(level int) int { return 5 * (level + 1) }

func (g *Game) finishGame(i int, result WinResult) {
	switch {
	case result.isDraw:
		g.scores.draws++
	case result.winner == PlayerX:
		g.scores.xWins++
	default:
		g.scores.oWins++
	}
	g.currency++
	g.recentResult = result
	g.recentTimer = resultTTL
	g.slots[i].result = result
	g.slots[i].animTimer = 1
}

// tryBuyUpgrade returns true if the click was consumed by a button area.
func (g *Game) tryBuyUpgrade(mx, my int) bool {
	if my < upgradeY || my >= upgradeY+btnH {
		return false
	}
	bx0 := btnStartX
	bx1 := btnStartX + btnW + btnGap
	bx2 := btnStartX + 2*(btnW+btnGap)

	switch {
	case mx >= bx0 && mx < bx0+btnW:
		cost := moreTicCost(g.moreTicLevel)
		if g.currency >= cost {
			g.currency -= cost
			g.moreTicLevel++
		}
		return true
	case mx >= bx1 && mx < bx1+btnW:
		cost := moreTacCost(g.moreTacLevel)
		if g.currency >= cost {
			g.currency -= cost
			g.moreTacLevel++
			g.slots = append(g.slots, boardSlot{board: NewBoard()})
		}
		return true
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

	// Advance per-slot animation timers.
	for i := range g.slots {
		s := &g.slots[i]
		if s.inAnim() {
			s.animTimer++
			if s.animTimer > animFrames {
				s.board.Reset()
				s.animTimer = 0
			}
		}
	}

	if g.recentTimer > 0 {
		g.recentTimer--
	}

	// Idle auto-moves: accumulate moreTicLevel per frame, fire one move per 60 accumulated.
	// This handles both sub-1/sec and multi-per-frame rates without a cap.
	if g.moreTicLevel > 0 {
		g.idleAccum += g.moreTicLevel
		for g.idleAccum >= 60 {
			g.idleAccum -= 60
			for i := range g.slots {
				if g.slots[i].inAnim() {
					continue
				}
				if g.slots[i].board.RandomMove() {
					if result, done := g.slots[i].board.CheckResult(); done {
						g.finishGame(i, result)
					}
				}
			}
		}
	}

	// Manual move: one move on a random non-animating board.
	if just {
		available := make([]int, 0, len(g.slots))
		for i := range g.slots {
			if !g.slots[i].inAnim() {
				available = append(available, i)
			}
		}
		if len(available) > 0 {
			i := available[rand.Intn(len(available))]
			if g.slots[i].board.RandomMove() {
				if result, done := g.slots[i].board.CheckResult(); done {
					g.finishGame(i, result)
				}
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colorBg)
	g.drawHeader(screen)
	g.drawBoards(screen)
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

// boardLayout returns grid dims and per-board pixel size for n boards.
func boardLayout(n int) (cols, rows, sz int) {
	if n < 1 {
		n = 1
	}
	cols = n
	if cols > 5 {
		cols = 5
	}
	rows = (n + cols - 1) / cols
	cellW := boardAreaW / cols
	cellH := boardAreaH / rows
	sz = cellW - 8
	if cellH-8 < sz {
		sz = cellH - 8
	}
	if sz < 1 {
		sz = 1
	}
	return
}

// boardOrigin returns the top-left pixel corner of board i.
func boardOrigin(i, cols, cellW, cellH, sz int) (x, y int) {
	row, col := i/cols, i%cols
	cx := boardAreaX + col*cellW + cellW/2
	cy := boardAreaY + row*cellH + cellH/2
	return cx - sz/2, cy - sz/2
}

func (g *Game) drawBoards(screen *ebiten.Image) {
	n := len(g.slots)
	cols, rows, sz := boardLayout(n)
	cellW := boardAreaW / cols
	cellH := boardAreaH / rows
	cs := float32(sz) / 3 // pixel size of one cell within a board

	for i := range g.slots {
		bx, by := boardOrigin(i, cols, cellW, cellH, sz)
		g.drawSingleBoard(screen, &g.slots[i], bx, by, sz, cs)
	}
}

func (g *Game) drawSingleBoard(screen *ebiten.Image, slot *boardSlot, bx, by, sz int, cs float32) {
	lw := float32(1.5)
	if sz >= 200 {
		lw = 2
	}

	vector.DrawFilledRect(screen, float32(bx), float32(by), float32(sz), float32(sz),
		color.RGBA{22, 22, 38, 255}, false)

	for i := 1; i < 3; i++ {
		x := float32(bx + i*sz/3)
		vector.StrokeLine(screen, x, float32(by), x, float32(by+sz), lw, colorGrid, false)
		y := float32(by + i*sz/3)
		vector.StrokeLine(screen, float32(bx), y, float32(bx+sz), y, lw, colorGrid, false)
	}
	vector.StrokeRect(screen, float32(bx), float32(by), float32(sz), float32(sz), lw, colorGrid, false)

	for pi, p := range slot.board.cells {
		row, col := pi/3, pi%3
		cx := float32(bx) + float32(col)*cs + cs/2
		cy := float32(by) + float32(row)*cs + cs/2
		switch p {
		case PlayerX:
			drawXPiece(screen, cx, cy, cs, colorX)
		case PlayerO:
			drawOPiece(screen, cx, cy, cs, colorO)
		}
	}

	if slot.inAnim() && !slot.result.isDraw {
		line := slot.result.line
		r0, c0 := line[0]/3, line[0]%3
		r2, c2 := line[2]/3, line[2]%3
		x0 := float32(bx) + float32(c0)*cs + cs/2
		y0 := float32(by) + float32(r0)*cs + cs/2
		x1 := float32(bx) + float32(c2)*cs + cs/2
		y1 := float32(by) + float32(r2)*cs + cs/2
		winColor := colorX
		if slot.result.winner == PlayerO {
			winColor = colorO
		}
		wlw := lw * 3
		if wlw < 3 {
			wlw = 3
		}
		vector.StrokeLine(screen, x0, y0, x1, y1, wlw, winColor, true)
	}
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	const resultY = 572.0
	const hintY = 602.0

	if g.recentTimer > 0 {
		var msg string
		var clr color.Color
		switch {
		case g.recentResult.isDraw:
			msg, clr = "DRAW", colorDim
		case g.recentResult.winner == PlayerX:
			msg, clr = "X WINS!", colorX
		default:
			msg, clr = "O WINS!", colorO
		}
		drawTC(screen, msg, g.face(32), screenW/2, resultY, clr)
	}

	if g.moreTicLevel == 0 {
		drawTC(screen, "SPACE or CLICK to play", g.face(14), screenW/2, hintY, colorDim)
	}
}

func (g *Game) drawUpgrades(screen *ebiten.Image) {
	drawTC(screen, "UPGRADES", g.face(13), screenW/2, float64(upgradeY)-14, colorDim)
	g.drawMoreTicBtn(screen)
	g.drawMoreTacBtn(screen)
	g.drawLockedBtn(screen, btnStartX+2*(btnW+btnGap), "MORE TOE")
}

func (g *Game) drawMoreTicBtn(screen *ebiten.Image) {
	x, y := float32(btnStartX), float32(upgradeY)
	cx, cy := float64(btnStartX+btnW/2), float64(upgradeY)

	cost := moreTicCost(g.moreTicLevel)
	canAfford := g.currency >= cost

	bgClr, borderClr := upgradeColors(canAfford)
	vector.DrawFilledRect(screen, x, y, float32(btnW), float32(btnH), bgClr, false)
	vector.StrokeRect(screen, x, y, float32(btnW), float32(btnH), 1.5, borderClr, false)

	drawTC(screen, "MORE TIC", g.face(14), cx, cy+16, colorWin)

	costClr := color.Color(colorDim)
	if canAfford {
		costClr = colorAfford
	}
	drawTC(screen, fmt.Sprintf("COST: %s", gamesLabel(cost)), g.face(11), cx, cy+36, costClr)
	if g.moreTicLevel == 0 {
		drawTC(screen, "unlock auto-move", g.face(11), cx, cy+56, colorDim)
	} else {
		drawTC(screen, fmt.Sprintf("LVL %d · %d/sec", g.moreTicLevel, g.moreTicLevel), g.face(11), cx, cy+56, colorDim)
	}
}

func (g *Game) drawMoreTacBtn(screen *ebiten.Image) {
	bx := btnStartX + btnW + btnGap
	x, y := float32(bx), float32(upgradeY)
	cx, cy := float64(bx+btnW/2), float64(upgradeY)

	cost := moreTacCost(g.moreTacLevel)
	canAfford := g.currency >= cost

	bgClr, borderClr := upgradeColors(canAfford)
	vector.DrawFilledRect(screen, x, y, float32(btnW), float32(btnH), bgClr, false)
	vector.StrokeRect(screen, x, y, float32(btnW), float32(btnH), 1.5, borderClr, false)

	drawTC(screen, "MORE TAC", g.face(14), cx, cy+16, colorWin)

	costClr := color.Color(colorDim)
	if canAfford {
		costClr = colorAfford
	}
	drawTC(screen, fmt.Sprintf("COST: %s", gamesLabel(cost)), g.face(11), cx, cy+36, costClr)
	drawTC(screen, fmt.Sprintf("+1 board (%d total)", len(g.slots)+1), g.face(11), cx, cy+56, colorDim)
}

func (g *Game) drawLockedBtn(screen *ebiten.Image, bx int, label string) {
	x, y := float32(bx), float32(upgradeY)
	cx, cy := float64(bx+btnW/2), float64(upgradeY)

	vector.DrawFilledRect(screen, x, y, float32(btnW), float32(btnH), colorBtnBg, false)
	vector.StrokeRect(screen, x, y, float32(btnW), float32(btnH), 1.5, colorBtnBorder, false)

	drawTC(screen, label, g.face(14), cx, cy+28, colorDim)
	drawTC(screen, "LOCKED", g.face(11), cx, cy+50, color.RGBA{70, 70, 100, 255})
}

func (g *Game) Layout(_, _ int) (int, int) { return screenW, screenH }

// upgradeColors returns bg and border colors based on affordability.
func upgradeColors(canAfford bool) (color.Color, color.Color) {
	if canAfford {
		return color.RGBA{20, 40, 25, 255}, colorAfford
	}
	return colorBtnBg, colorBtnBorder
}

// gamesLabel formats a game count with correct singular/plural.
func gamesLabel(n int) string {
	if n == 1 {
		return "1 game"
	}
	return fmt.Sprintf("%d games", n)
}

func drawXPiece(screen *ebiten.Image, cx, cy, cs float32, clr color.Color) {
	pad := cs * 0.28
	half := cs / 2
	lw := cs / 10
	if lw < 2 {
		lw = 2
	}
	vector.StrokeLine(screen, cx-half+pad, cy-half+pad, cx+half-pad, cy+half-pad, lw, clr, true)
	vector.StrokeLine(screen, cx+half-pad, cy-half+pad, cx-half+pad, cy+half-pad, lw, clr, true)
}

func drawOPiece(screen *ebiten.Image, cx, cy, cs float32, clr color.Color) {
	r := cs/2 - cs*0.15
	lw := cs / 10
	if lw < 2 {
		lw = 2
	}
	vector.StrokeCircle(screen, cx, cy, r, lw, clr, true)
}

func drawTC(screen *ebiten.Image, str string, face *text.GoTextFace, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, face, op)
}
