package main

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"strconv"
	"tic-tac-toe/game"
)

func (ws *webServer) getAllGames(ctx *fasthttp.RequestCtx) {
	logger := ws.Log.WithFields(logrus.Fields{"req": strconv.FormatUint(ctx.ID(), 26), "f": "getAllGames"})

	games, err := ws.storage.ListRaw()
	if err != nil {
		logger.Errorln(err)
		ctx.SetStatusCode(err.(*game.GameError).Status)
		return
	}

	ctx.SetContentType(applicationJson)
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
		_, err := w.Write([]byte{'['})
		if err != nil {
			logger.Errorln("can't write response:", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		for idx, g := range games {
			_, err = w.Write(g)
			if err != nil {
				logger.Errorln("can't write response:", err)
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				return
			}
			if idx < len(games)-1 {
				_, err = w.Write([]byte{','})
			}
		}
		_, err = w.Write([]byte{']'})
		if err != nil {
			logger.Errorln("can't write response:", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
		err = w.Flush()
		if err != nil {
			logger.Errorln("can't flush response:", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}
	})
}

func (ws *webServer) startNewGame(ctx *fasthttp.RequestCtx) {
	logger := ws.Log.WithFields(logrus.Fields{"req": strconv.FormatUint(ctx.ID(), 26), "f": "startNewGame"})

	body := ctx.Request.Body()
	logger.Debugln("BODY:", string(body))

	p := ws.parserPool.Get()
	val, err := p.ParseBytes(body)
	if err != nil {
		logger.Errorln("can't parse request:", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	board := val.GetStringBytes("board")
	if board == nil {
		logger.Errorln("can't parse board")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	ws.parserPool.Put(p)

	var g *game.Game
	var userSign byte

	switch game.WhoMovesFirst(board) {
	case game.ComputerMove:
		// computer always plays X
		userSign = game.GambleSign()[0]
	case game.XChar:
		// user is playing X
		userSign = game.XChar
	case game.OChar:
		// user is playing O
		userSign = game.OChar
	default:
		logger.Errorln("invalid first board: game: %+v", g)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	// create game and make move
	g = game.NewGame(board, userSign)
	g.MakeMove()

	logger.Debugln("game: %+v", g)

	// save game
	err = ws.storage.Save(g)
	if err != nil {
		logger.Errorln("can't save new game:", err)
		ctx.SetStatusCode(err.(*game.GameError).Status)
		return
	}

	// response to user with location
	ctx.SetContentType(applicationJson)
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetBody([]byte(`{"location":"https://` + ws.Addr + `/api/v1/games/` + g.Id() + `"}`))
}

func (ws *webServer) getGame(ctx *fasthttp.RequestCtx) {
	logger := ws.Log.WithFields(logrus.Fields{"req": strconv.FormatUint(ctx.ID(), 26), "f": "getGame"})

	gameId := ctx.UserValue("game_id").(string)
	logger.Debugln("game_id:", gameId)
	logger.Infoln("game_id:", gameId)

	if !ws.storage.IsValidGameId(gameId) {
		logger.Errorln("getGame: invalid game id", gameId)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	g, err := ws.storage.GetRaw(gameId)
	if err != nil {
		logger.Errorln("getGame:", err)
		ctx.SetStatusCode(err.(*game.GameError).Status)
		return
	}

	setOkResponse(ctx, g)
}

func (ws *webServer) makeMove(ctx *fasthttp.RequestCtx) {
	logger := ws.Log.WithFields(logrus.Fields{"req": strconv.FormatUint(ctx.ID(), 26), "f": "makeMove"})
	gameId := ctx.UserValue("game_id").(string)
	logger.Debugln("game_id:", gameId)

	if !ws.storage.IsValidGameId(gameId) {
		logger.Errorln("invalid game id", gameId)
		setReason(ctx, game.NewGameError(fasthttp.StatusBadRequest, "invalid game id"))
		return
	}

	p := ws.parserPool.Get()
	val, err := p.ParseBytes(ctx.Request.Body())
	if err != nil {
		logger.Errorln("makeMove: can't parse request:", err)
		setReason(ctx, game.NewGameError(fasthttp.StatusBadRequest, "can't parse request", err))
		return
	}

	board := val.GetStringBytes("board")
	if board == nil {
		logger.Errorln("makeMove: can't parse board")
		setReason(ctx, game.NewGameError(fasthttp.StatusBadRequest, "can't get board from request"))
		return
	}
	ws.parserPool.Put(p)

	g, err := ws.storage.Get(gameId)
	if err != nil {
		logger.Errorln("makeMove:", err)
		setReason(ctx, err.(*game.GameError))
		return
	}

	// check game status
	if g.Status() != game.RUNNING {
		logger.Errorln("makeMove: game already finished with status", g.Status())
		setReason(ctx, game.NewGameError(fasthttp.StatusBadRequest, "game already finished with status "+g.Status()))
		return
	}

	// validate user move
	if ok, err := g.SetNewBoard(board); !ok {
		logger.Errorln("makeMove: board is invalid")
		setReason(ctx, err.(*game.GameError))
		return
	}

	// check is user a winner?
	if g.CheckWin(g.UserSign()) == game.RUNNING {
		// game continue
		g.MakeMove()
		// check is computer a winner?
		g.CheckWin(g.CompSign())
	}

	// save game
	err = ws.storage.Save(g)
	if err != nil {
		logger.Errorln("makeMove: can't save game:", err)
		setReason(ctx, err.(*game.GameError))
		return
	}

	// marshal game and send to user
	setOkResponse(ctx, g.Marshal())
}

func (ws *webServer) deleteGame(ctx *fasthttp.RequestCtx) {
	logger := ws.Log.WithFields(logrus.Fields{"req": strconv.FormatUint(ctx.ID(), 26), "f": "deleteGame"})
	gameId := ctx.UserValue("game_id").(string)
	logger.Debugln("game_id:", gameId)

	if !ws.storage.IsValidGameId(gameId) {
		logger.Errorln("invalid game id", gameId)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err := ws.storage.Delete(gameId)
	if err != nil {
		logger.Errorln(err)
		ctx.SetStatusCode(err.(*game.GameError).Status)
		return
	}

	setOkResponse(ctx, nil)
}

func setReason(ctx *fasthttp.RequestCtx, err *game.GameError) {
	ctx.SetStatusCode(err.Status)
	ctx.SetContentType(applicationJson)
	if err.Status == fasthttp.StatusInternalServerError {
		ctx.SetBody([]byte(`{"reason":"internal server error"}`))
	} else {
		ctx.SetBody([]byte(fmt.Sprintf(`{"reason":%q}`, err.Error())))
	}
}

func setOkResponse(ctx *fasthttp.RequestCtx, res []byte) {
	ctx.SetContentType(applicationJson)
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(res)
}
