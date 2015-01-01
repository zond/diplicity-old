package meta

type GameState string

const (
	GameStateCreated GameState = "Created"
	GameStateStarted GameState = "Started"
	GameStateEnded   GameState = "Ended"
)

var GameStates = map[string]GameState{
	"Created": GameStateCreated,
	"Started": GameStateStarted,
	"Ended":   GameStateEnded,
}
