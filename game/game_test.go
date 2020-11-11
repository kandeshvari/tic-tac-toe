package game

import "testing"

type boards struct {
	oldBoard string
	newBoard string
	userSign int
}

type checkWin struct {
	board  string
	status string
	char   byte
}

var (
	suite1Bad = []boards{
		{"---------", "---------", UserSetX},
		{"---------", "---------", UserSetO},
		{"---------", "---------X", UserSetX},
		{"---------", "--------X-", UserSetX},
		{"---------", "X---------", UserSetX},
		{"---------", "-X--------", UserSetX},
		{"---------", "X-------X", UserSetX},
		{"---------", "O-------X", UserSetX},
		{"---------", "O-------X", UserSetO},
		{"---------", "----XO---", UserSetX},
		{"---------", "--X_-----", UserSetX},
		{"---------", "---=-----", UserSetX},
		{"---------", "---=-----", UserSetO},
		{"X--------", "-XO------", UserSetO},
		{"X--------", "-XO------", UserSetX},
		{"---------", "XXXXXXXXX", UserSetX},
		{"---------", "XXXXOXXXX", UserSetX},
		{"---------", "XXXX-XXXX", UserSetX},
	}

	suite1Ok = []boards{
		{"---------", "X--------", UserSetX},
		{"---------", "-X-------", UserSetX},
		{"---------", "--------X", UserSetX},
		{"O--------", "OX-------", UserSetX},
		{"O--------", "O-------X", UserSetX},
		{"--------O", "X-------O", UserSetX},
		{"--------O", "----X---O", UserSetX},
		{"----O----", "X---O----", UserSetX},
		{"----O----", "----O---X", UserSetX},
		{"----O----", "----OX---", UserSetX},
		{"----O----", "---XO----", UserSetX},
		{"-O--XX---", "-O--XX--O", UserSetO},
	}

	suite2Computer = []string{
		`---------`,
	}
	suite2UserX = []string{
		`--X------`,
		`X--------`,
		`--------X`,
	}
	suite2UserO = []string{
		`----O----`,
		`O--------`,
		`--------O`,
	}
	suite2Bad = []string{
		`--------`,
		`----------`,
		`---XO----`,
		`---OO----`,
		`---XX----`,
		`OO-------`,
		`O-------O`,
		`-------OO`,
		`X-------X`,
		`XX------X`,
		`XXXXXXXXX`,
		`OOOOOOOOO`,
	}

	suite3 = []checkWin{
		{`---------`, RUNNING, XChar},
		{`---------`, RUNNING, OChar},
		{`XXOOO-XOX`, RUNNING, OChar},
		{`OOXXX-OXO`, RUNNING, XChar},
		{`XOOXOXOX-`, RUNNING, XChar},
		{`X-O-X-XOO`, RUNNING, XChar},
		{`X-OXX-XOO`, XWON, XChar},
		{`XOX-XOO-X`, XWON, XChar},
		{`XXXO-O--O`, XWON, XChar},
		{`OX-OOXXXO`, OWON, OChar},
		{`OOXXXOOXO`, DRAW, OChar},
		{`XXOOOXXOX`, DRAW, XChar},
	}
)

func TestGame_SetNewBoard(t *testing.T) {
	for idx, s := range suite1Bad {
		firstLetter := "a"
		if s.userSign == UserSetO {
			firstLetter = "f"
		}

		g := &Game{
			id:     firstLetter,
			board:  []byte(s.oldBoard),
			status: RUNNING,
		}

		if ok, err := g.SetNewBoard([]byte(s.newBoard)); ok {
			t.Fatalf("bad input (%d) became valid: %v: %s", idx, s, err)
		}
	}

	for idx, s := range suite1Ok {
		firstLetter := "a"
		if s.userSign == UserSetO {
			firstLetter = "f"
		}

		g := &Game{
			id:     firstLetter,
			board:  []byte(s.oldBoard),
			status: RUNNING,
		}

		if ok, err := g.SetNewBoard([]byte(s.newBoard)); !ok {
			t.Fatalf("valid input (%d) became bad: %v: %s", idx, s, err)
		}
	}
}

func TestWhoMovesFirst(t *testing.T) {
	for _, s := range suite2Computer {
		x := WhoMovesFirst([]byte(s))
		if x != ComputerMove {
			t.Fatalf("board (%s) await ComputerMove(%d) got %v", s, ComputerMove, x)
		}
	}

	for _, s := range suite2UserO {
		x := WhoMovesFirst([]byte(s))
		if x != OChar {
			t.Fatalf("board (%s) await OChar(%d) got %v", s, OChar, x)
		}
	}

	for _, s := range suite2UserX {
		x := WhoMovesFirst([]byte(s))
		if x != XChar {
			t.Fatalf("board (%s) await UserSetX(%d) got %v", s, XChar, x)
		}
	}

	for _, s := range suite2Bad {
		x := WhoMovesFirst([]byte(s))
		if x == OChar || x == XChar || x == ComputerMove {
			t.Fatalf("board (%s) is bad but got %v", s, x)
		}
	}
}

func TestCheckWin(t *testing.T) {
	for _, s := range suite3 {
		firstLetter := "a"
		if s.char == OChar {
			firstLetter = "f"
		}

		g := &Game{
			id:     firstLetter,
			board:  []byte(s.board),
			status: s.status,
		}

		win := g.CheckWin(s.char)
		if win != s.status {
			t.Fatalf("board (%s) status is %s but got %s", s.board, s.status, win)
		}
	}
}
