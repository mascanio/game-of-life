package drawboard

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	vertexShaderSource = `
    #version 420
    in vec3 vp;
    void main() {
        gl_Position = vec4(vp, 1.0);
    }
` + "\x00"

	fragmentShaderSource = `
    #version 420
    out vec4 frag_colour;
    void main() {
        frag_colour = vec4(1, 1, 1, 1);
    }
` + "\x00"
)

type points []float32

type square struct {
	x, y float32
}

type drawBoard struct {
	cells   [][]points
	vaos    [][]uint32
	xrows   int
	yrows   int
	xsize   int
	ysize   int
	window  *glfw.Window
	program uint32
}

func MakeDrawBoard(xrows, yrows, xsize, ysize int) drawBoard {
	if xsize != ysize {
		panic("xsize and ysize must be equal")
	}
	squareSize := 2 / float32(xrows)
	r := drawBoard{
		cells: make([][]points, xrows),
		vaos:  make([][]uint32, xrows),
		xrows: xrows,
		yrows: yrows,
		xsize: xsize,
		ysize: ysize,
	}
	r.window = initGlfw(&r)
	r.program = initOpenGL()
	for i := range r.cells {
		r.cells[i] = make([]points, yrows)
		r.vaos[i] = make([]uint32, yrows)
		xpos := i - xrows/2
		xnorm := float32(xpos) / float32(xrows/2)
		for j := range r.cells[i] {
			ypos := j - yrows/2
			ynorm := float32(ypos) / float32(yrows/2)
			r.cells[i][j] = squareGetPoints(square{
				x: xnorm,
				y: ynorm,
			}, float32(squareSize))
			r.vaos[i][j] = makeVAO(r.cells[i][j])
		}
	}
	return r
}

func DrawboardTerminate(drawBoard *drawBoard) {
	drawBoard.window.Destroy()
	glfw.Terminate()
}

func DrawCell(drawBoard *drawBoard, x, y int) {
	gl.BindVertexArray(drawBoard.vaos[x][y])
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func DrawCellsIf(drawBoard *drawBoard, cells [][]bool) {
	for i := range cells {
		for j := range cells[i] {
			if cells[i][j] {
				DrawCell(drawBoard, i, j)
			}
		}
	}
}

func DrawIteration(drawBoard *drawBoard, cells [][]bool) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(drawBoard.program)
	DrawCellsIf(drawBoard, cells)
	drawBoard.window.SwapBuffers()
}

func ShouldClose(drawBoard *drawBoard) bool {
	return drawBoard.window.ShouldClose()
}

func makeVAO(points points) uint32 {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Create a vertex buffer object
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	// Copy our points into the currently bound vertex buffer.
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	// Specify the layout of the vertex data
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	return vao
}

func squareGetPoints(s square, size float32) points {
	return points{
		s.x, s.y, 0,
		s.x + size, s.y, 0,
		s.x + size, s.y + size, 0,

		s.x, s.y, 0,
		s.x, s.y + size, 0,
		s.x + size, s.y + size, 0,
	}
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw(drawBoard *drawBoard) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(drawBoard.xsize, drawBoard.ysize, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// initOpenGL initializes OpenGL and returns an intiialized program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}
