package game

// game error wrapper to save status code
type GameError struct {
	Status  int
	message string
}

func (e *GameError) Error() string {
	return e.message
}

func NewGameError(status int, text string, err ...error) *GameError {
	var errMsg string

	for _, e := range err {
		errMsg += ": " + e.Error()
	}

	return &GameError{
		Status:  status,
		message: text + errMsg,
	}
}
