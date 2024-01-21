package main

import (
	"fmt"
	"runtime"
	"time"

	draw "github.com/mascanio/game-of-life/drawboard"
)

const (
	xrows  = 32
	yrows  = 32
	width  = 720
	height = 720
	fps    = 60
)

func main() {
	runtime.LockOSThread()
	drawBoard := draw.MakeDrawBoard(xrows, yrows, width, height)
	defer draw.DrawboardTerminate(&drawBoard)

	cells := make([][]bool, xrows)
	for i := range cells {
		cells[i] = make([]bool, yrows)
		for j := range cells[i] {
			cells[i][j] = false
		}
		if i == 0 {
			cells[i][0] = true
			cells[i][1] = true
			cells[i][2] = true
		}
	}

	for !draw.ShouldClose(&drawBoard) {
		fpsCurrentTime := time.Now()
		draw.DrawIteration(&drawBoard, cells)
		fpsTime := time.Since(fpsCurrentTime)
		fmt.Println("FPS:", 1/fpsTime.Seconds())
		time.Sleep(time.Second / fps)
	}
}
