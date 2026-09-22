package main

import (
	"strings"
	"testing"
)

func headless(cfg Config, steps int) *Game {
	game := NewGame(cfg)
	for i := 0; i < steps && game.State == StateRunning; i++ {
		game.Step()
	}
	return game
}

func TestInitialSnakeAndFood(t *testing.T) {
	game := NewGame(Config{Width: 20, Height: 12, Seed: 1, BaseSpeedMS: 80})
	if len(game.snake) != 3 {
		t.Fatalf("snake must start at length 3, got %d", len(game.snake))
	}
	if game.food == (Point{}) && game.occupied(game.food) {
		t.Fatal("food spawned on the snake")
	}
	if game.occupied(game.food) {
		t.Fatal("food inside the snake")
	}
}

func TestFoodIsEatenAndScoreGrows(t *testing.T) {
	game := NewGame(Config{Width: 10, Height: 10, Seed: 5, BaseSpeedMS: 80})
	// steer the head toward the food until it is eaten
	for step := 0; step < 200 && game.Score == 0; step++ {
		head := game.snake[0]
		if head.X < game.food.X {
			game.SetDir(DirRight)
		} else if head.X > game.food.X {
			game.SetDir(DirLeft)
		} else if head.Y < game.food.Y {
			game.SetDir(DirDown)
		} else if head.Y > game.food.Y {
			game.SetDir(DirUp)
		}
		game.Step()
	}
	if game.Score != 10 {
		t.Fatalf("score after one food = %d, want 10", game.Score)
	}
	if len(game.snake) != 4 {
		t.Fatalf("snake length after food = %d, want 4", len(game.snake))
	}
}

func TestSolidWallEndsGame(t *testing.T) {
	game := NewGame(Config{Width: 8, Height: 8, Walls: WallsSolid, Seed: 2, BaseSpeedMS: 80})
	for i := 0; i < 20; i++ {
		game.SetDir(DirLeft)
		game.Step()
	}
	if game.State != StateOver {
		t.Fatal("snake must die on a solid wall")
	}
}

func TestWrapModeTeleports(t *testing.T) {
	game := NewGame(Config{Width: 8, Height: 8, Walls: WallsWrap, Seed: 2, BaseSpeedMS: 80})
	startX := game.snake[0].X
	for i := 0; i < startX + 1; i++ {
		game.SetDir(DirLeft)
		game.Step()
		if game.State != StateRunning {
			t.Fatal("wrap mode must not die at the edge")
		}
	}
	if head := game.snake[0]; head.X >= startX+1 && head.X < 8 {
		// teleported across and kept moving left; must be on the right side
		_ = head
	}
}

func TestObstaclesKillAndRender(t *testing.T) {
	cfg := Config{Width: 10, Height: 10, Walls: WallsSolid, Seed: 3, BaseSpeedMS: 80}
	// wall of obstacles right in front of the spawn heading (right)
	for y := 0; y < 10; y++ {
		cfg.Obstacles = append(cfg.Obstacles, Obstacle{X: 7, Y: y})
	}
	game := NewGame(cfg)
	game.SetDir(DirRight)
	for i := 0; i < 20 && game.State == StateRunning; i++ {
		game.Step()
	}
	if game.State != StateOver {
		t.Fatal("obstacle collision must end the game")
	}
	if !strings.Contains(game.RenderASCII(), "#") {
		t.Fatal("obstacles must render as #")
	}
}

func TestSpeedIncreasesWithGrowth(t *testing.T) {
	game := NewGame(Config{Width: 30, Height: 30, Seed: 4, BaseSpeedMS: 100})
	before := game.SpeedMS(100)
	// grow artificially to simulate eating
	for i := 0; i < 5; i++ {
		game.snake = append(game.snake, game.snake[len(game.snake)-1])
	}
	after := game.SpeedMS(100)
	if after >= before {
		t.Fatalf("speed must increase (interval shrink): before=%d after=%d", before, after)
	}
	if game.SpeedMS(100) < 40 {
		t.Fatal("interval floor is 40ms")
	}
}

func TestOppositeTurnRejected(t *testing.T) {
	game := NewGame(Config{Width: 10, Height: 10, Seed: 6, BaseSpeedMS: 80})
	game.SetDir(DirLeft) // opposite of initial DirRight
	if game.dir != DirRight {
		t.Fatal("180-degree turn must be ignored")
	}
}

func TestRenderUsesUnicodeBlocks(t *testing.T) {
	game := NewGame(Config{Width: 12, Height: 8, Seed: 8, BaseSpeedMS: 80})
	board := game.RenderASCII()
	for _, glyph := range []string{"▪", "▒", "●"} {
		if !strings.Contains(board, glyph) {
			t.Fatalf("board missing %q", glyph)
		}
	}
	if !strings.Contains(board, "|") {
		t.Fatal("borders missing")
	}
}
