package main

import (
	"fmt"
	"math/rand"
)

// Game modes for walls.
type WallMode int

const (
	WallsWrap WallMode = iota
	WallsSolid
)

type Point struct {
	X, Y int
}

type Direction int

const (
	DirUp Direction = iota
	DirRight
	DirDown
	DirLeft
)

// Opposite returns the reverse direction; used to reject 180° turns.
func (d Direction) Opposite() Direction {
	switch d {
	case DirUp:
		return DirDown
	case DirDown:
		return DirUp
	case DirRight:
		return DirLeft
	default:
		return DirRight
	}
}

// Obstacle is a static wall cell (mode: obstacles).
type Obstacle Point

type Config struct {
	Width, Height int
	Walls         WallMode
	Obstacles     []Obstacle
	Seed          int64
	BaseSpeedMS   int // starting tick interval; <=0 means 80
}

type GameState int

const (
	StateRunning GameState = iota
	StateOver
)

type Game struct {
	cfg       Config
	snake     []Point // head first
	dir       Direction
	food      Point
	obstacles map[Point]bool
	rng       *rand.Rand
	Score     int
	State     GameState
}

func NewGame(cfg Config) *Game {
	if cfg.Width < 5 {
		cfg.Width = 20
	}
	if cfg.Height < 5 {
		cfg.Height = 12
	}
	g := &Game{
		cfg:       cfg,
		dir:       DirRight,
		obstacles: map[Point]bool{},
		rng:       rand.New(rand.NewSource(cfg.Seed)),
	}
	head := Point{cfg.Width / 2, cfg.Height / 2}
	g.snake = []Point{head, {head.X - 1, head.Y}, {head.X - 2, head.Y}}
	for _, o := range cfg.Obstacles {
		g.obstacles[Point(o)] = true
	}
	g.placeFood()
	return g
}

func (g *Game) occupied(p Point) bool {
	for _, seg := range g.snake {
		if seg == p {
			return true
		}
	}
	return g.obstacles[p]
}

// placeFood picks a random free cell; deterministic under the seed.
func (g *Game) placeFood() {
	free := make([]Point, 0, g.cfg.Width*g.cfg.Height)
	for y := 0; y < g.cfg.Height; y++ {
		for x := 0; x < g.cfg.Width; x++ {
			p := Point{x, y}
			if !g.occupied(p) {
				free = append(free, p)
			}
		}
	}
	if len(free) == 0 {
		return // board full: win condition territory
	}
	g.food = free[g.rng.Intn(len(free))]
}

// SetDir queues a direction change; 180° turns are ignored.
func (g *Game) SetDir(d Direction) {
	if d.Opposite() != g.dir {
		g.dir = d
	}
}

// Step advances one tick. Returns true when food was eaten.
func (g *Game) Step() bool {
	if g.State != StateRunning {
		return false
	}
	head := g.snake[0]
	next := head
	switch g.dir {
	case DirUp:
		next.Y--
	case DirDown:
		next.Y++
	case DirLeft:
		next.X--
	case DirRight:
		next.X++
	}

	switch g.cfg.Walls {
	case WallsWrap:
		next.X = (next.X + g.cfg.Width) % g.cfg.Width
		next.Y = (next.Y + g.cfg.Height) % g.cfg.Height
	default: // solid
		if next.X < 0 || next.Y < 0 || next.X >= g.cfg.Width || next.Y >= g.cfg.Height {
			g.State = StateOver
			return false
		}
	}
	if g.obstacles[next] {
		g.State = StateOver
		return false
	}

	ate := next == g.food
	if !ate {
		g.snake = g.snake[:len(g.snake)-1] // tail moves away unless growing
	}
	// colliding with the remaining body ends the game
	for _, seg := range g.snake {
		if seg == next {
			g.State = StateOver
			return false
		}
	}
	g.snake = append([]Point{next}, g.snake...)
	if ate {
		g.Score += 10
		if len(g.snake) < g.cfg.Width*g.cfg.Height {
			g.placeFood()
		}
		return true
	}
	return false
}

// SpeedMS returns the tick interval for the current score: the snake
// speeds up as it grows.
func (g *Game) SpeedMS(base int) int {
	speedup := len(g.snake) * 8
	ms := base - speedup
	if ms < 40 {
		ms = 40
	}
	return ms
}

// RenderASCII draws the board as unicode blocks (head ▪, body ▒, food ●,
// walls #).
func (g *Game) RenderASCII() string {
	grid := make([][]rune, g.cfg.Height)
	for y := range grid {
		grid[y] = make([]rune, g.cfg.Width)
		for x := range grid[y] {
			grid[y][x] = ' '
		}
	}
	for _, o := range g.cfg.Obstacles {
		grid[o.Y][o.X] = '#'
	}
	for i, seg := range g.snake {
		if i == 0 {
			grid[seg.Y][seg.X] = '▪'
		} else {
			grid[seg.Y][seg.X] = '▒'
		}
	}
	grid[g.food.Y][g.food.X] = '●'
	out := ""
	for _, row := range grid {
		out += fmt.Sprintf("|%s|\n", string(row))
	}
	return out
}
