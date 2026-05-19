package main

import "math/rand"

type Player int8

const (
	Empty   Player = 0
	PlayerX Player = 1
	PlayerO Player = 2
)

var winLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

type Board struct {
	cells [9]Player
	turn  Player
}

type WinResult struct {
	winner Player
	line   [3]int
	isDraw bool
}

func NewBoard() Board {
	return Board{turn: PlayerX}
}

func (b *Board) RandomMove() bool {
	var empty []int
	for i, c := range b.cells {
		if c == Empty {
			empty = append(empty, i)
		}
	}
	if len(empty) == 0 {
		return false
	}
	b.cells[empty[rand.Intn(len(empty))]] = b.turn
	b.turn = opponent(b.turn)
	return true
}

func (b *Board) CheckResult() (WinResult, bool) {
	for _, line := range winLines {
		a, bb, c := line[0], line[1], line[2]
		if b.cells[a] != Empty && b.cells[a] == b.cells[bb] && b.cells[a] == b.cells[c] {
			return WinResult{winner: b.cells[a], line: line}, true
		}
	}
	for _, c := range b.cells {
		if c == Empty {
			return WinResult{}, false
		}
	}
	return WinResult{isDraw: true}, true
}

func (b *Board) Reset() {
	*b = NewBoard()
}

func opponent(p Player) Player {
	if p == PlayerX {
		return PlayerO
	}
	return PlayerX
}
