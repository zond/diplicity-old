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

type EndReason string

const (
	BeforePhaseType   dip.PhaseType = "Before"
	AfterPhaseType    dip.PhaseType = "After"
	DuringPhaseType   dip.PhaseType = "During"
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
	"Created": GameStateCreated,
	"Started": GameStateStarted,
	"Ended":   GameStateEnded,
}
