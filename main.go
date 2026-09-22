package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const version = "1.0.0"

const usage = `snake-tui v%s — classic snake in unicode blocks

USAGE
  snake-tui play [-m mode] [-o count] [-s seed] [-speed ms]
  snake-tui (-h | --help | -V | --version)

MODES
  wrap      edges teleport (default)
  solid     hitting a wall ends the game

OPTIONS
  -o <count>    number of static obstacles (default 0)
  -s <seed>     food/obstacle seed (default: clock)
  -speed <ms>   starting tick interval, gets faster as you grow (80)

KEYS  h/l left-right, k/j up-down, q quit
Exit codes: 0 ok, 2 usage error.`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			fmt.Printf(usage+"\n", version)
			return
		case "-V", "--version":
			fmt.Println("snake-tui", version)
			return
		}
	}

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "play" {
		args = args[1:] // optional subcommand; flags follow
	}
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	mode := fs.String("m", "wrap", "wall mode: wrap | solid")
	obstacles := fs.Int("o", 0, "number of static obstacles")
	seed := fs.Int64("s", 0, "seed (0 = clock)")
	speed := fs.Int("speed", 80, "starting tick interval in ms")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "snake-tui: unexpected argument %q\n", fs.Arg(0))
		os.Exit(2)
	}
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	cfg := Config{Width: 32, Height: 16, Seed: *seed, BaseSpeedMS: *speed}
	switch *mode {
	case "solid":
		cfg.Walls = WallsSolid
	case "wrap":
		cfg.Walls = WallsWrap
	default:
		fmt.Fprintf(os.Stderr, "snake-tui: unknown mode %q (wrap | solid)\n", *mode)
		os.Exit(2)
	}
	// obstacles: random cells in the inner area, never near the spawn row
	rng := rand.New(rand.NewSource(*seed))
	for i := 0; i < *obstacles; i++ {
		cfg.Obstacles = append(cfg.Obstacles, Obstacle{
			X: 3 + rng.Intn(cfg.Width-6),
			Y: 2 + rng.Intn(cfg.Height-4),
		})
	}

	if os.Getenv("SNAKE_HEADLESS") != "" {
		runHeadless(cfg)
		return
	}
	runInteractive(cfg)
}

// runHeadless plays a few dozen deterministic steps printing the final
// board — used by the test suite to verify rendering without a tty.
func runHeadless(cfg Config) {
	game := NewGame(cfg)
	turns := []Direction{DirDown, DirRight, DirRight, DirUp, DirRight, DirDown}
	for step := 0; step < 60 && game.State == StateRunning; step++ {
		game.SetDir(turns[step%len(turns)])
		game.Step()
	}
	fmt.Print(game.RenderASCII())
	fmt.Printf("score=%d state=%d len=%d\n", game.Score, game.State, len(game.snake))
}

func runInteractive(cfg Config) {
	game := NewGame(cfg)

	raw := enableRawMode()
	defer raw()
	fmt.Print("\033[?25l\033[H\033[2J")

	tickerInterval := time.Duration(game.SpeedMS(cfg.BaseSpeedMS)) * time.Millisecond
	ticker := time.NewTicker(tickerInterval)
	input := make(chan byte, 8)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				continue
			}
			input <- buf[0]
		}
	}()

	for {
		select {
		case key := <-input:
			switch key {
			case 'h':
				game.SetDir(DirLeft)
			case 'l':
				game.SetDir(DirRight)
			case 'k':
				game.SetDir(DirUp)
			case 'j':
				game.SetDir(DirDown)
			case 'q':
				fmt.Print("\033[?25h")
				fmt.Printf("bye — score %d\n", game.Score)
				return
			}
		case <-ticker.C:
			game.Step()
			fmt.Print("\033[H")
			fmt.Print(game.RenderASCII())
			fmt.Printf("score %d | length %d\n", game.Score, len(game.snake))
			if game.State == StateOver {
				fmt.Print("\033[?25h")
				fmt.Printf("GAME OVER — score %d\n", game.Score)
				return
			}
			ticker.Reset(time.Duration(game.SpeedMS(cfg.BaseSpeedMS)) * time.Millisecond)
		}
	}
}
