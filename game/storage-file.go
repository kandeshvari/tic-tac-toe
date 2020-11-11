package game

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	maxFileSize = 512    // maximum file size for storage one game
	backupExt   = ".bak" // backup file extenstion
)

type StorageFile struct {
	path          string
	gameIdPattern *regexp.Regexp
	log           *log.Logger
	rwm           sync.RWMutex
	parserPool    *fastjson.ParserPool
}

func NewStorage(path string, logger *log.Logger) (*StorageFile, error) {
	// check could we use this dir as storage
	tmpFile, err := ioutil.TempFile(path, "test")
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't create test storage file", err)
	}
	err = tmpFile.Close()
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't close test storage file", err)
	}
	err = os.Remove(tmpFile.Name())
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't remove test storage file", err)
	}

	p, err := regexp.Compile("^[0-9a-f-]{36}$")
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't compile game id pattern", err)
	}

	return &StorageFile{
		path:          path,
		gameIdPattern: p,
		log:           logger,
		parserPool:    &fastjson.ParserPool{},
	}, nil
}

func (s *StorageFile) GetRaw(gameId string) ([]byte, error) {
	fname := s.path + "/" + gameId
	s.rwm.RLock()
	if fInfo, err := os.Stat(fname); err == nil {
		// check if file size more than `maxFileSize`. It could prevent DoS via reading large files
		if fInfo.Size() > maxFileSize {
			return nil, NewGameError(fasthttp.StatusInternalServerError, "game file too large")
		}
	} else if os.IsNotExist(err) {
		return nil, NewGameError(fasthttp.StatusNotFound, "game not exists", err)
	} else {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "error while checking file", err)
	}

	content, err := ioutil.ReadFile(fname)
	s.rwm.RUnlock()

	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't read file content", err)
	}
	return content, nil
}

func (s *StorageFile) Get(gameId string) (*Game, error) {

	content, err := s.GetRaw(gameId)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "file content has zero size")
	}

	p := s.parserPool.Get()
	game, err := Unmarshal(p, content)
	if err != nil {
		return nil, err
	}
	s.parserPool.Put(p)

	return game, nil
}

func (s *StorageFile) Save(game *Game) error {
	removeBackup := false
	fname := s.path + "/" + game.id

	/* make old game backup */
	if ok, _ := s.IsGameExists(game.id); ok {
		s.rwm.Lock()
		err := os.Rename(fname, fname+backupExt)
		s.rwm.Unlock()
		if err != nil {
			return NewGameError(fasthttp.StatusInternalServerError, "can't create game file backup", err)
		}
		removeBackup = true
	}

	buf := game.Marshal()

	/* save new game file */
	s.rwm.Lock()
	err := ioutil.WriteFile(fname, buf, 0640)
	s.rwm.Unlock()
	if err != nil {
		return NewGameError(fasthttp.StatusInternalServerError, "can't write game file", err)
	}

	/* remove backup */
	if removeBackup {
		s.rwm.Lock()
		err = os.Remove(fname + backupExt)
		s.rwm.Unlock()
		if err != nil {
			s.log.Printf("game %s: can't remove backup file %s: %s", game.id, fname+backupExt, err)
		}
	}

	return nil
}

func (s *StorageFile) List() ([]*Game, error) {
	s.rwm.RLock()
	files, err := ioutil.ReadDir(s.path)
	s.rwm.RUnlock()
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't read storage dir", err)
	}
	res := make([]*Game, 0, len(files))
	for _, f := range files {
		if !s.IsValidGameId(f.Name()) {
			s.log.Printf("invalid game id (%s) detected in storage. Remove it manually", f.Name())
			continue
		}
		game, err := s.Get(f.Name())
		if err != nil {
			return nil, err
		}
		res = append(res, game)
	}
	return res, nil
}

func (s *StorageFile) ListRaw() ([][]byte, error) {
	s.rwm.RLock()
	files, err := ioutil.ReadDir(s.path)
	s.rwm.RUnlock()
	if err != nil {
		return nil, NewGameError(fasthttp.StatusInternalServerError, "can't read storage dir", err)
	}
	res := make([][]byte, 0, len(files))
	for _, f := range files {
		if !s.IsValidGameId(f.Name()) {
			s.log.Printf("invalid game id (%s) detected in storage. Remove it manually", f.Name())
			continue
		}
		game, err := s.GetRaw(f.Name())
		if err != nil {
			return nil, err
		}
		res = append(res, game)
	}
	return res, nil
}

func (s *StorageFile) IsValidGameId(gameId string) bool {
	return s.gameIdPattern.MatchString(gameId)
}

func (s *StorageFile) IsGameExists(gameId string) (bool, error) {
	fname := s.path + "/" + gameId
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	if _, err := os.Stat(fname); err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func (s *StorageFile) Delete(gameId string) error {
	if ok, _ := s.IsGameExists(gameId); !ok {
		return NewGameError(fasthttp.StatusNotFound, "game not found")
	}

	// double check, just in case
	if !s.IsValidGameId(gameId) {
		return NewGameError(fasthttp.StatusBadRequest, "invalid game id")
	}

	fname := s.path + "/" + gameId
	s.rwm.Lock()
	err := os.Remove(fname)
	s.rwm.Unlock()
	if err != nil {
		return NewGameError(fasthttp.StatusInternalServerError, "can't remove game file", err)
	}

	return nil
}

func (s *StorageFile) Shutdown() error {
	// sleep a second to wait all read/write storage operations will be done
	time.Sleep(1 * time.Second)

	return nil
}
