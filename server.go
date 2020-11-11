package main

import (
	"github.com/fasthttp/router"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
	"github.com/valyala/fastjson"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tic-tac-toe/game"
)

const applicationJson = "application/json"

type webServer struct {
	Addr       string
	Log        *log.Logger
	ln         net.Listener
	router     *router.Router
	debug      bool
	wg         sync.WaitGroup
	certFile   string
	keyFile    string
	storage    game.Storage
	parserPool *fastjson.ParserPool // reuse parsers to avoid memory allocations
	server     *fasthttp.Server
}

func NewServer(addr string, certFile, key string, storage game.Storage, logger *log.Logger) *webServer {
	s := &webServer{
		Addr:       addr,
		Log:        logger,
		router:     router.New(),
		debug:      true,
		certFile:   certFile,
		keyFile:    key,
		storage:    storage,
		parserPool: &fastjson.ParserPool{},
	}
	return s
}

func (ws *webServer) Close() {
	_ = ws.ln.Close()
}

func (ws *webServer) Shutdown() {
	//ws.Close() // close listener
	ws.Log.Info("shutting down web server")

	err := ws.server.Shutdown()
	if err != nil {
		ws.Log.Errorln("http server shutdown error:", err)
	}

	err = ws.storage.Shutdown()
	if err != nil {
		ws.Log.Errorln("storage shutdown error:", err)
	}

	// we can add more comprehensive logic here, if we will use other types of storage.
}

func (ws *webServer) Run() (err error) {
	ws.registerHandlers()

	// reuse port to run per-core server instance
	ws.ln, err = reuseport.Listen("tcp4", ws.Addr)
	if err != nil {
		return err
	}

	ws.server = &fasthttp.Server{
		Handler:            ws.router.Handler,
		Name:               "tic-tac-toe server",
		ReadBufferSize:     1024,
		MaxConnsPerIP:      1024,
		MaxRequestsPerConn: 128,
		MaxRequestBodySize: 1024,
		Logger:             ws.Log,
	}

	// install shutdown handler
	killChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(killChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-killChan
		ws.Log.Infof("received shutdown signal: %s", sig)
		ws.Shutdown()
		done <- true
	}()

	// start server
	ws.Log.Info("starting server at ", *addr)
	err = ws.server.ServeTLS(ws.ln, ws.certFile, ws.keyFile)
	if err != nil {
		ws.Log.Errorln("can't start server:", err)
		return err
	}

	// wait while service is shutting down
	<-done
	ws.Log.Info("service shutdown")

	return nil
}

func (ws *webServer) registerHandlers() {
	ws.router.GET("/api/v1/games", ws.Recovery(ws.getAllGames))
	ws.router.POST("/api/v1/games", ws.Recovery(ws.startNewGame))
	ws.router.GET("/api/v1/games/{game_id}", ws.Recovery(ws.getGame))
	ws.router.PUT("/api/v1/games/{game_id}", ws.Recovery(ws.makeMove))
	ws.router.DELETE("/api/v1/games/{game_id}", ws.Recovery(ws.deleteGame))
}

func (ws *webServer) Recovery(next func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	fn := func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if rvr := recover(); rvr != nil {
				ws.Log.Errorln("recover:", rvr)
				ctx.Error("recover", fasthttp.StatusInternalServerError)
			}
		}()

		// do next
		next(ctx)
	}
	return fn
}
