# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tic-tac-idle: an idler-style tic-tac-toe game written in Go using the Ebiten v2 game engine, compiled to WASM and hosted in-browser.

## Build & Run

```bash
# Run locally (native window)
go run .

# Build for WASM (browser target)
GOOS=js GOARCH=wasm go build -o web/main.wasm .

# Copy the Go WASM runtime shim (do this once or after Go upgrades)
cp $(go env GOROOT)/misc/wasm/wasm_exec.js web/

# Serve the web/ directory locally to test WASM
npx serve web/
# or
python3 -m http.server 8080 --directory web/
```

Use `go test ./...` to run tests and `go vet ./...` to lint.

## Ebiten Notes

- Target ebiten v2 only; use non-deprecated APIs.
- Assets (fonts, images) must be embedded or sourced from packages bundled with ebiten — no external asset downloads.
- Entry point is `ebiten.RunGame(&Game{})` in `main.go`; the `Game` struct implements `ebiten.Game` (Update, Draw, Layout).

## Architecture (Planned)

The game has three layers:

**Game state** (`game/` or top-level structs): board state, move history, win/draw/loss counters, upgrade levels. All mutation goes through a single `State` struct so the idle loop and player input share one source of truth.

**Idle engine**: a ticker (driven by `Update()` calls at 60 TPS) that fires automatic moves based on "more tic" upgrade level. Timing uses frame counts rather than `time.Now()` to stay deterministic.

**Upgrade system**: three upgrades from the project charter —
- *more tic* — auto-move rate (1–3/sec); costs games played
- *more tac* — board multiplier; each additional board shares the auto-move ticker
- *more toe* — cosmetic, buyable 3–5 times; purchasing all copies ends the game

**Multi-board layout**: when *more tac* is purchased, boards render side-by-side. Manual clicks randomly target one board; auto-moves distribute across all boards.

## Deployment

GitHub Actions builds the WASM artifact and pushes to GitHub Pages on a subdomain of the personal website. The workflow lives in `.github/workflows/`. DNS propagation is the long pole — set up the subdomain record early.

## Game Loop Summary

```
Update() each frame:
  handle input (click / spacebar → random move on a random board)
  advance idle ticker → fire auto-moves at upgrade rate
  check win/draw → record result, reset board(s), increment currency

Draw() each frame:
  render board grid(s)
  render score/upgrade UI
```
