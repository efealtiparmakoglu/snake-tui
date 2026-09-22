# snake-tui

Classic snake in your terminal — **unicode block rendering**, two wall
modes, static obstacles and a speed curve that tightens as you grow.
Pure Go stdlib (raw mode via portable `stty`).

```text
$ snake-tui play -m solid -o 4 -s 5
|                                |
|          ●                     |
|                   #        #   |
|           #             ▒      |
|           #             ▪#     |
score 0 | length 3
```

## Features

- **Wall modes** — `wrap` teleports across edges; `solid` ends the game.
- **Obstacles** — `-o N` scatters N static `#` blocks (never on the
  spawn row); crashing into one is fatal.
- **Speed curve** — tick interval starts at `-speed` ms and shrinks 8 ms
  per body segment, floored at 40 ms.
- **Unicode rendering** — head `▪`, body `▒`, food `●`, obstacles `#`.
- Deterministic `-s seed` for food and obstacles; 180° turns are
  rejected; the core game logic is separate from the terminal layer and
  fully unit-tested (`SNAKE_HEADLESS=1` renders without a tty).

## Keys

`h`/`l` left/right · `k`/`j` up/down · `q` quit

## Usage

```
snake-tui play [-m wrap|solid] [-o obstacles] [-s seed] [-speed ms]
```

## Development

```console
$ go test -count=1 .     # eating, walls, obstacles, speed curve, rendering
$ go test -race -count=1 .
$ gofmt -l . && go vet .
```

## License

MIT — see [LICENSE](LICENSE).
