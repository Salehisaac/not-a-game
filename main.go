package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
	"math/rand"
)

const (
	WALL = 1
	NOTHING = 0
	PLAYER = 69
	MAX_SAMPLES = 100
	POINT = 911
	BOMB = 666
)

type input struct{
	pressedKey byte

}


func (i *input) update() {
	b := make([]byte, 1)
	go func() {
		os.Stdin.SetReadDeadline(time.Now().Add(time.Millisecond * 16))
		os.Stdin.Read(b)   
		i.pressedKey = b[0]
	}()
}
		   


type level struct{
	width, height int
	data [][]int
}

type position struct{
	x, y int
}

type point struct{
	pos position
	score int
}

type bomb struct{
	pos position
	level *level
	speedCounter int
	speed        int
}

type player struct{
	pos position
	level *level
	input *input
	score int
	
}

func (p *player) update(){

	if p.input.pressedKey == 97 {
		p.pos.x -= 1
	}

	if p.input.pressedKey == 100 {
		p.pos.x += 1
	}
		
	if p.pos.x == p.level.width-2{
		p.pos.x = 1
	}
	if p.pos.x == 0{
		p.pos.x = p.level.width-2
	}

	if p.input.pressedKey == 119 {
		p.pos.y -= 1
		time.Sleep(time.Millisecond * 16)
	}

	if p.input.pressedKey == 115 {
		p.pos.y += 1
		time.Sleep(time.Millisecond * 16)
	}
		
	if p.pos.y == p.level.height-1 {
		p.pos.y = 1
	}
	if p.pos.y == 0{
		p.pos.y = p.level.height-2
	}
}



type stats struct{
	start time.Time
	frames int
	fps float64
}

func NewState() *stats{
	return &stats{
		start: time.Now(),
	}
}

func (s *stats) update(){
	s.frames++
	if s.frames == MAX_SAMPLES{
		s.fps = float64(s.frames) / time.Since(s.start).Seconds()
		s.frames = 0
		s.start = time.Now()
	}
}

func (b *bomb) update() {
	if b.pos.y < b.level.height-2 {
		b.pos.y += 1
	}
}
func newLevel(width, height int) *level {
	data := make([][]int, height)

	for h := 0; h < height; h++ {
		for w := 0; w < width; w++ {
			data[h] = make([]int, width)
		}
	}
	for h := 0; h < height; h++ {
		for w := 0; w < width; w++ {
			if w == width-1 || h == height-1 || h == 0 || w == 0 {
				data[h][w] = WALL
			}
		}
	}
	return &level{
		width: width,
		height: height,
		data : data,
	}
}

func (g *game) bombLoop(b *bomb) {
	for {
		time.Sleep(time.Millisecond * 25)
		b.speedCounter++
		if b.speedCounter >= 10 {
			b.speedCounter = 0
			g.level.set(b.pos, NOTHING)
			b.update()
			g.level.set(b.pos, BOMB)

			if b.pos.y >= g.level.height-2 {
				g.level.set(b.pos, NOTHING)
				b.pos.y = 1
				source := rand.NewSource(time.Now().UnixNano())
				r := rand.New(source)
				b.pos.x = r.Intn(g.level.width-2) + 1
			}
		}
	}
}

  type game struct{
	isRunning bool
	level 	*level
	drawBuf *bytes.Buffer
	stats   *stats
	player *player
	input *input
	point *point
	bombs []*bomb
	bombCount int
	delayedUpdate chan struct{}
}

func NewGame(width, height int) *game{
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	level := newLevel(width, height)
	inpu := &input{}
	g := &game{
		level: level,
		drawBuf: new(bytes.Buffer),
		stats: NewState(),
		input: inpu,
		bombCount: 3,
		player : &player{
			level: level,
			pos : position{
				x : 2,
				y: 5,
			},
			input: inpu,
		},
		point : &point{
			pos : position{
				x: 2,
				y : 6,
			},
			score: 10,
		},
		bombs: make([]*bomb, 0),
		delayedUpdate: make(chan struct{}, 1),
	}

	for i := 0; i < g.bombCount; i++ { 
		b := &bomb{
			level: g.level,
			pos: position{
				x: rand.Intn(g.level.width-2) + 1,
				y: 1,
			},
			speedCounter: i * 10, 
		}
		g.bombs = append(g.bombs, b)
		go g.bombLoop(b)
	}
	return g
}

func (l *level) set(pos position, v int){
	l.data[pos.y][pos.x] = v
}

func(g *game) start(){
	g.isRunning = true
	g.loop()
}
func(g *game) loop(){

	for g.isRunning{
		g.input.update()
		g.update()
		g.render()
		g.stats.update()
		time.Sleep(time.Millisecond * 25)//limit fps

	}

	g.renderGameOver()
}

func (g *game) renderGameOver() {
	g.drawBuf.Reset()
	fmt.Fprint(os.Stdout, "\033[2J\033[1;1H") // Clear the screen
	message := `
***************************
*                         *
*       GAME OVER         *
*                         *
***************************
`
	g.drawBuf.WriteString(message)
	fmt.Fprint(os.Stdout, g.drawBuf.String())
}

func(g *game) render(){
	g.drawBuf.Reset()
	fmt.Fprint(os.Stdout,"\033[2J\033[1:1H")
	g.renderLevel()
	g.renderStates()
	fmt.Fprint(os.Stdout, g.drawBuf.String())
}

func(g *game) randPosintion(pos position) position{
	g.level.set(pos, NOTHING)
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	pos.x = r.Intn(g.level.width-2) + 1
	pos.y = r.Intn(g.level.height-2) + 1
	return pos
}

func(g *game) addBomb(){
	b := &bomb{
		level: g.level,
		pos: position{
			x: rand.Intn(g.level.width-2) + 1,
			y: 1,
		},
		speedCounter: 4 * 10, 
	}
	g.bombs = append(g.bombs, b)
	go g.bombLoop(b)
}
func (g *game) update() {
	if g.point.pos.x == g.player.pos.x && g.point.pos.y == g.player.pos.y {
		g.point.pos = g.randPosintion(g.point.pos)
		g.player.score += g.point.score
		if g.player.score > 50 {
			g.addBomb()
		}
	}
	for _ , bomb :=range g.bombs {
		if bomb.pos.x == g.player.pos.x && bomb.pos.y == g.player.pos.y {
			g.isRunning = false
		}
	}


	g.level.set(g.player.pos, NOTHING)
	g.player.update()
	g.level.set(g.player.pos, PLAYER)

	for _, b := range g.bombs {
		g.level.set(b.pos, BOMB)
	}
	
	time.AfterFunc(2 * time.Second, func() {
		g.delayedUpdate <- struct{}{}
	})

	select {
	case <-g.delayedUpdate:
		g.level.set(g.point.pos, POINT)
	default:
		// No update needed
	}
}

func (g *game) renderStates(){
	g.drawBuf.WriteString("--STATS\n")
	g.drawBuf.WriteString(fmt.Sprintf("FPS: %.2f\n", g.stats.fps))
	g.drawBuf.WriteString(fmt.Sprintf("POINT: %v\n", g.player.score))
}



func (g *game) renderLevel(){
	
	for h := 0; h < g.level.height; h++ {
		for w := 0; w < g.level.width; w++ {
			if g.level.data[h][w] == NOTHING{
				g.drawBuf.WriteString(" ")
			}
			if g.level.data[h][w] == WALL{
				g.drawBuf.WriteString("🌫")
			}
			if g.level.data[h][w] == PLAYER{
				g.drawBuf.WriteString("🛩")
			}
			if g.level.data[h][w] == POINT{
				g.drawBuf.WriteString("⛴")
			}
			if g.level.data[h][w] == BOMB{
				g.drawBuf.WriteString("֍")
			}
		}
		g.drawBuf.WriteString("\n")
	}
	fmt.Println(g.drawBuf.String())
}

func main() {
	width := 80
	height := 40
	g := NewGame(width, height)
	g.start()
}