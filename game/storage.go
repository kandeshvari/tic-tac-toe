package game

type Storage interface {
	Get(gameId string) (*Game, error)
	GetRaw(gameId string) ([]byte, error)
	List() ([]*Game, error)
	ListRaw() ([][]byte, error)
	Save(game *Game) error
	Delete(gameId string) error
	Shutdown() error

	IsValidGameId(gameId string) bool
	IsGameExists(gameId string) (bool, error)
}
