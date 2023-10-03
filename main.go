package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type (
	field      int8
	userAction int8
)

const (
	fieldEmpty = field(iota)
	fieldSnake
	fieldFood
)

const (
	userActionUp = userAction(iota)
	userActionLeft
	userActionRight
	userActionDown
)

type point2d struct {
	x, y int
}

func (c point2d) isValid(state *gameState) bool {
	return c.x < int(state.boardWidth) && c.y < int(state.boardHeight) && c.x >= 0 && c.y >= 0
}

func (c point2d) isEqual(c2 point2d) bool {
	return c.x == c2.x && c.y == c2.y
}

type snakeField struct {
	pos       point2d
	nextField *snakeField
}

func (s *snakeField) last() *snakeField {
	if s.nextField == nil {
		return s
	}
	return s.nextField.last()
}

func (s *snakeField) beforeLast() *snakeField {
	if s.nextField == nil {
		return nil
	}
	if s.nextField.nextField == nil {
		return s
	}
	return s.nextField.beforeLast()
}

func (s snakeField) count() uint {
	current := &s
	var i uint
	for i = 0; current != nil; i++ {
		current = current.nextField
	}
	return i
}

type gameState struct {
	// The 2D board of the game
	board       [][]field // [x][y]int8
	boardWidth  uint
	boardHeight uint

	round uint64
	score uint64

	desiredSnakeLen uint // includes the head
	snakeHead       *snakeField

	userAction userAction

	// context for breaking the game loop
	appCtx       context.Context
	appCtxCancel func()
}

// based on https://github.com/nsf/termbox-go/blob/8f994c032445f01d32d27544f62d4dfeaecfd52b/_demos/hello_world.go#L25
func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	lineX := x
	for _, c := range msg {
		if c == '\n' {
			y++
			lineX = x
		}
		termbox.SetCell(lineX, y, c, fg, bg)
		lineX += runewidth.RuneWidth(c)
	}
}

// getField returns type of field located at the set 2d point.
// Returns field, and true if field exists, or false if it does not.
func (gs *gameState) getField(p point2d) (field, bool) {
	if !p.isValid(gs) {
		return field(0), false
	}

	// check for snake fields
	currentField := gs.snakeHead
	for currentField != nil {
		if currentField.pos.isEqual(p) {
			return fieldSnake, true
		}
		currentField = currentField.nextField
	}

	// check for regular fields
	return gs.board[p.x][p.y], true
}

func (gs *gameState) randPoint2d() point2d {
	c := point2d{}
	c.x = rand.Intn(int(gs.boardWidth))
	c.y = rand.Intn(int(gs.boardHeight))
	return c
}

// print prints the game into the terminal according to the gameState
func (gs *gameState) print() {
	bgColor := termbox.Attribute(0)
	// print the fields
	for x := uint(0); x < gs.boardWidth*2; x += 2 {
		for y := uint(0); y < gs.boardHeight; y++ {
			switch gs.board[x/2][y] {
			case fieldEmpty:
				bgColor = termbox.ColorDarkGray
			case fieldFood:
				bgColor = termbox.ColorRed
			}
			termbox.SetCell(int(x), int(y), ' ', termbox.ColorDefault, bgColor)
			termbox.SetCell(int(x+1), int(y), ' ', termbox.ColorDefault, bgColor)
		}
	}

	// print the snakefields
	field := gs.snakeHead
	head := true
	for field != nil {
		if head {
			bgColor = termbox.ColorLightGreen
			head = false
		} else {
			bgColor = termbox.ColorGreen
		}
		termbox.SetCell(field.pos.x*2, field.pos.y, '*', termbox.ColorDefault, bgColor)
		termbox.SetCell((field.pos.x*2)+1, field.pos.y, '*', termbox.ColorDefault, bgColor)
		field = field.nextField
	}

	// print score, etc.
	footer := fmt.Sprintf("Round: %v\n\rScore: %v\n\rSnake: length %v, head x %v, head y %v  ",
		gs.round, gs.score, gs.desiredSnakeLen, gs.snakeHead.pos.x, gs.snakeHead.pos.y)
	tbprint(0, int(gs.boardHeight), termbox.ColorDefault, termbox.ColorDefault, footer)

	err := termbox.Flush()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

}

// init inits the gameState. Should be called before doing any operations with gameState.
func (gs *gameState) init(boardWidth uint, boardHeight uint) {
	gs.board = make([][]field, boardWidth, boardWidth)
	for i := uint(0); i < boardWidth; i++ {
		gs.board[i] = make([]field, boardHeight, boardHeight)
	}

	gs.boardWidth = boardWidth
	gs.boardHeight = boardHeight

	gs.appCtx, gs.appCtxCancel = context.WithCancel(context.Background())
}

// keyPress listens to new key presses and changes gameState's userAction accordingly
func (gs *gameState) keyPressListener() {
	defer gs.appCtxCancel()
	for {
		select {
		case <-gs.appCtx.Done():
			return
		default:
			switch event := termbox.PollEvent(); event.Type {
			case termbox.EventKey:
				switch event.Key {
				case termbox.KeyArrowUp:
                    if gs.userAction != userActionDown {
                        gs.userAction = userActionUp
                    }
				case termbox.KeyArrowLeft:
                    if gs.userAction != userActionRight {
                        gs.userAction = userActionLeft
                    }
				case termbox.KeyArrowRight:
                    if gs.userAction != userActionLeft {
                        gs.userAction = userActionRight
                    }
				case termbox.KeyArrowDown:
                    if gs.userAction != userActionUp {
                        gs.userAction = userActionDown
                    }
				case termbox.KeyCtrlC, termbox.KeyCtrlZ, termbox.KeyCtrlX:
					return
				}
			case termbox.EventInterrupt:
				return
			case termbox.EventError:
				os.Exit(1)
			}

		}

	}
}

// newRound() moves the snake one step according to the user input
func (gs *gameState) newRound() {
	gs.round++

	// move the head forward
	newSnakeHead := snakeField{
		pos:       gs.snakeHead.pos,
		nextField: gs.snakeHead,
	}

	switch gs.userAction {
	case userActionUp:
		newSnakeHead.pos.y--
		if newSnakeHead.pos.y == -1 {
			newSnakeHead.pos.y = int(gs.boardHeight) - 1
		}
	case userActionLeft:
		newSnakeHead.pos.x--
		if newSnakeHead.pos.x == -1 {
			newSnakeHead.pos.x = int(gs.boardWidth) - 1
		}
	case userActionRight:
		newSnakeHead.pos.x++
		if newSnakeHead.pos.x == int(gs.boardWidth) {
			newSnakeHead.pos.x = 0
		}
	case userActionDown:
		newSnakeHead.pos.y++
		if newSnakeHead.pos.y == int(gs.boardHeight) {
			newSnakeHead.pos.y = 0
		}
	}

	// prevent nil pointer exceptions
	if !newSnakeHead.pos.isValid(gs) {
		return
	}

	// end the game if the snake ever tries to bite itself
	if field, _ := gs.getField(newSnakeHead.pos); field == fieldSnake {
		gs.appCtxCancel()
		return
	}
	gs.snakeHead = &newSnakeHead

	// if we just hit food with the snake head, pick it up and create new food
	if gs.board[gs.snakeHead.pos.x][gs.snakeHead.pos.y] == fieldFood {
		gs.board[gs.snakeHead.pos.x][gs.snakeHead.pos.y] = fieldEmpty
		gs.score++
		gs.desiredSnakeLen++
		gs.newFood()
	}

	if gs.snakeHead.count() > gs.desiredSnakeLen {
		// move the tail forward
		beforeLastField := gs.snakeHead.beforeLast()
		beforeLastField.nextField = nil
	}
}

func (gs *gameState) newFood() {
	p := gs.randPoint2d()
	field, exists := gs.getField(p)
	if !exists {
		log.Fatalf("Generated invalid location for food, %v", field)
		return
	}

	if field == fieldSnake {
		gs.newFood()
		return
	}

	gs.board[p.x][p.y] = fieldFood
}

func menu() *gameState {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("\nTERMSNAKE\nPlease enter the desired width of the board: ")
	scanner.Scan()
	desiredWidth, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	fmt.Printf("Please enter the desired height of the board: ")
	scanner.Scan()
	desiredHeight, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}

	// check if the terminal is big enough
	err = termbox.Init()
	if err != nil {
		log.Fatalf("%v\n", err.Error())
	}
	termWidth, termHeight := termbox.Size()

	if desiredWidth < 5 || desiredHeight < 5 {
		fmt.Printf("The minimum dimensions are 5x5!")
		os.Exit(3)
	}

	if desiredWidth > termWidth/2 || desiredHeight > termHeight-3 {
		fmt.Printf("Error: the desired dimensions are too large for your terminal!\n"+
			"Your terminal size: %vx%v\n"+
			"Your desired size: %vx%v\n", termWidth/2, termHeight-3, desiredWidth, desiredHeight)
		os.Exit(2)
	}

	state := gameState{}
	state.init(uint(desiredWidth), uint(desiredHeight))
	return &state
}

func main() {
	state := menu()

	// create the snake
	body2 := snakeField{
		pos:       point2d{int(state.boardWidth / 2), int((state.boardHeight / 2) + 1)},
		nextField: nil,
	}
	body1 := snakeField{
		pos:       point2d{int(state.boardWidth / 2), int(state.boardHeight / 2)},
		nextField: &body2,
	}
	state.snakeHead = &snakeField{
		pos:       point2d{int(state.boardWidth / 2), int((state.boardHeight / 2) - 1)},
		nextField: &body1,
	}
	state.desiredSnakeLen = 3

	go state.keyPressListener()

	// main game loop
	t := time.NewTicker(150 * time.Millisecond)
	state.newFood()
	for {
		select {
		case <-t.C:
			state.newRound()
			state.print()
		case <-state.appCtx.Done():
			goto exit
		}
	}

	// make sure we leave the terminal "clean"
exit:
	termbox.Close()
	fmt.Println("Game over!")
	fmt.Printf("Round: %v\tScore: %v\tSnake: length %v, head x %v, head y %v\n",
		state.round, state.score, state.desiredSnakeLen, state.snakeHead.pos.x, state.snakeHead.pos.y)
	os.Exit(0)
}
