package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	draw "github.com/mascanio/game-of-life/drawboard"
)

const (
	xrows  = 128
	yrows  = 128
	width  = 720
	height = 720
	fps    = 1000
)

type game struct {
	game    [][][]bool
	current int
	rules   []rule
}

type rule func(g *game, x, y int) bool

func isCellAlive(g *game, x, y int) bool {
	return g.game[g.current][x][y]
}

func countAliveNeighbors(g *game, x, y, dist int) int {
	aliveNeighbors := 0
	for i := x - dist; i <= x+dist; i++ {
		if i >= 0 && i < xrows {
			for j := y - dist; j <= y+dist; j++ {
				if j >= 0 && j < yrows && !(i == x && j == y) && isCellAlive(g, i, j) {
					aliveNeighbors++
				}
			}
		}
	}
	return aliveNeighbors
}

func countAliveNeighbors3x3(g *game, x, y int) int {
	return countAliveNeighbors(g, x, y, 1)
}

func born(g *game, x, y int) bool {
	if isCellAlive(g, x, y) {
		return true
	}
	return countAliveNeighbors3x3(g, x, y) == 3
}

func die(g *game, x, y int) bool {
	if !isCellAlive(g, x, y) {
		return false
	}
	aliveNeighbors := countAliveNeighbors3x3(g, x, y)
	return !(aliveNeighbors < 2 || aliveNeighbors > 3)
}

func GameSwapBoard(g *game) {
	g.current = (g.current + 1) % 2
}

func GameGetCells(g *game) [][]bool {
	return g.game[g.current]
}

func GameSetCellRandom(g *game, x, y int) {
	// generate random number 0 or 1
	g.game[g.current][x][y] = rand.Intn(2) == 1
}

func GameSetCellsRandom(g *game) {
	for i := range g.game[g.current] {
		for j := range g.game[g.current][i] {
			GameSetCellRandom(g, i, j)
		}
	}
}

func MakeGame(xrows, yrows int) game {
	gameBoard := make([][][]bool, 2)
	for i := range gameBoard {
		gameBoard[i] = make([][]bool, xrows)
		for j := range gameBoard[i] {
			gameBoard[i][j] = make([]bool, yrows)
		}
	}
	return game{gameBoard, 0, []rule{born, die}}
}

func setNextGenCell(g *game, x, y int) {
	nextGenIdx := (g.current + 1) % 2
	if isCellAlive(g, x, y) {
		g.game[nextGenIdx][x][y] = die(g, x, y)
	} else {
		g.game[nextGenIdx][x][y] = born(g, x, y)
	}
}

func NewGeneration(g *game) {
	for i := range g.game[g.current] {
		for j := range g.game[g.current][i] {
			setNextGenCell(g, i, j)
		}
	}
	// Swap current and next generation
	g.current = (g.current + 1) % 2
}

func main() {
	runtime.LockOSThread()
	drawBoard := draw.MakeDrawBoard(xrows, yrows, width, height)
	defer draw.DrawboardTerminate(&drawBoard)

	game := MakeGame(xrows, yrows)
	GameSetCellsRandom(&game)

	generations := 0
	for !draw.ShouldClose(&drawBoard) {
		fpsCurrentTime := time.Now()
		draw.DrawIteration(&drawBoard, GameGetCells(&game))
		fpsTime := time.Since(fpsCurrentTime)
		fmt.Printf("FPS: %v, frame time: %v\n", 1/fpsTime.Seconds(), fpsTime)
		gameTime := time.Now()
		NewGeneration(&game)
		fmt.Printf("game time: %v\n", time.Since(gameTime))
		time.Sleep(time.Second / fps)
		generations++
		fmt.Printf("generations: %v\n", generations)
	}
}
