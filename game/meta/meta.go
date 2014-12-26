package meta

import (
	"fmt"

	dip "github.com/zond/godip/common"
)

type Consequence int

const (
	ReliabilityHit Consequence = 1 << iota
	NoWait
	Surrender
)

var Consequences = map[string]Consequence{
	"ReliabilityHit": ReliabilityHit,
	"NoWait":         NoWait,
	"Surrender":      Surrender,
}

type ChatFlag int

const (
	ChatPrivate = 1 << iota
	ChatGroup
	ChatConference
)

var ChatFlags = map[string]ChatFlag{
	"Private":    ChatPrivate,
	"Group":      ChatGroup,
	"Conference": ChatConference,
}

type EndReason string

const (
	BeforePhaseType   dip.PhaseType = "Before"
	AfterPhaseType    dip.PhaseType = "After"
	Anonymous         dip.Nation    = "Anonymous"
	ZeroActiveMembers EndReason     = "ZeroActiveMembers"
)

func SoloVictory(n dip.Nation) EndReason {
	return EndReason(fmt.Sprintf("SoloVictory:%v", n))
}

type GameState int

const (
	GameStateCreated GameState = iota
	GameStateStarted
	GameStateEnded
)

var GameStates = map[string]GameState{
	"created": GameStateCreated,
	"started": GameStateStarted,
	"ended":   GameStateEnded,
}

type SecretFlag int

const (
	SecretBeforeGame = 1 << iota
	SecretDuringGame
	SecretAfterGame
)

var SecretFlags = map[string]SecretFlag{
	"BeforeGame": SecretBeforeGame,
	"DuringGame": SecretDuringGame,
	"AfterGame":  SecretAfterGame,
}
