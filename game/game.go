package game

import (
	"bytes"
	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"math/rand"
	"time"
)

const (
	RUNNING = "RUNNING"
	XWON    = "X_WON"
	OWON    = "O_WON"
	DRAW    = "DRAW"
)

const (
	// chars
	DashChar = 45
	XChar    = 88
	OChar    = 79

	// first move constants
	ComputerMove = 0

	// user moves
	UserSetO = 98
	UserSetX = 117
)

type Game struct {
	id     string
	board  []byte
	status string
}

func NewGame(board []byte, userSign byte) *Game {
	// user plays X -> first uuid letter == a
	// user plays O -> first uuid letter == f
	firstLetter := "a"
	if userSign == OChar {
		firstLetter = "f"
	}

	return &Game{
		id:     firstLetter + uuid.NewV4().String()[1:],
		board:  board,
		status: RUNNING,
	}
}

// get user sign
func (g *Game) UserSign() byte {
	if g.id[0] == 'a' {
		return XChar
	}
	return OChar
}

// get computer sign
func (g *Game) CompSign() byte {
	if g.id[0] == 'a' {
		return OChar
	}
	return XChar
}

// make computer's move
func (g *Game) MakeMove() {
	var compSign byte = XChar
	if g.UserSign() == XChar {
		compSign = OChar
	}

	// AI :)
	rand.Seed(time.Now().UnixNano())
	for _, idx := range rand.Perm(9) {
		if g.board[idx] == DashChar {
			g.board[idx] = compSign
			break
		}
	}
}

// create json string from Game struct
func (g *Game) Marshal() []byte {
	return []byte(`{"id":"` + g.id + `","board":"` + string(g.board) + `","status":"` + g.status + `"}`)
}

func (g *Game) Id() string {
	return g.id
}

func (g *Game) Status() string {
	return g.status
}

func (g *Game) Board() []byte {
	return g.board
}

// compare previous board with new one and validate user move
func (g *Game) SetNewBoard(newBoard []byte) (bool, error) {
	sum := 0
	if len(newBoard) != 9 {
		return false, NewGameError(fasthttp.StatusBadRequest, "invalid board length")
	}
	for i := 0; i < 9; i++ {
		// oldBoard is always valid
		switch newBoard[i] {
		case DashChar, XChar, OChar:
			// '-' xor 'X' == 117
			// '-' xor 'O' == 98
			// '-' xor '-' == 0
			sum += int(g.board[i] ^ newBoard[i])
		default:
			// board contains invalid chars
			return false, NewGameError(fasthttp.StatusBadRequest, "board contains invalid chars")
		}
	}

	// check is user has made correct move
	if (sum == UserSetX && g.UserSign() == XChar) || (sum == UserSetO && g.UserSign() == OChar) {
		// update board
		g.board = newBoard

		return true, nil
	}

	// board sign didn't match user's first move. strange
	return false, NewGameError(fasthttp.StatusBadRequest, "move not valid")
}

// check winner
func (g *Game) CheckWin(s byte) string {
	coords := [][]int{
		{0, 4, 8},
		{2, 4, 6},
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
	}

	// check WIN position
	for _, c := range coords {
		if g.board[c[0]] == s && g.board[c[1]] == s && g.board[c[2]] == s {
			if s == XChar {
				g.status = XWON
				return XWON
			} else if s == OChar {
				g.status = OWON
				return OWON
			}
		}
	}

	// check DRAW
	numDash := false
	for i := 0; i < 9; i++ {
		if g.board[i] == DashChar {
			numDash = true
			break
		}
	}
	if !numDash {
		g.status = DRAW
		return DRAW
	}

	// no winners, continue
	return RUNNING
}

// parse game file to Game struct
func Unmarshal(p *fastjson.Parser, buf []byte) (*Game, error) {
	val, err := p.ParseBytes(buf)
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't parse game file", err)
	}

	return &Game{
		id:     string(val.GetStringBytes("id")),
		board:  val.GetStringBytes("board"),
		status: string(val.GetStringBytes("status")),
	}, nil

}

// Realize who moves first. If user - return the user's sign UserSetO or UserSetX.
// If return value is not in ComputerMove, UserSetO or UserSetX - user has sent invalid board
func WhoMovesFirst(board []byte) int {
	if len(board) != 9 {
		return -1
	}

	if bytes.Equal(board, []byte("---------")) {
		return ComputerMove
	}

	var sum int
	for _, b := range board {
		switch int(b) {
		case XChar, OChar:
			sum += int(b)
		case DashChar:
			continue
		default:
			return -1
		}
	}
	return sum
}

// gamble player signs
func GambleSign() [2]byte {
	rand.Seed(time.Now().UnixNano())
	chars := [2]byte{XChar, OChar}
	rand.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	return chars
}
