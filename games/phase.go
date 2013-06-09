package games

import (
	"appengine"
	"appengine/datastore"
	"common"
	"fmt"
	dip "github.com/zond/godip/common"
)

func latestPhaseByGameIdKey(k *datastore.Key) string {
	return fmt.Sprintf("%v{Game:%v,Latest}", phaseKind, k)
}

type Phases []*Phase

type Phase struct {
	Id      *datastore.Key `json:"id" datastore:"-"`
	Season  dip.Season     `json:"season"`
	Year    int            `json:"year"`
	Type    dip.PhaseType  `json:"type"`
	Ordinal int            `json:"ordinal"`
}

func (self *Phase) Save(c appengine.Context, gameId *datastore.Key) *Phase {
	latest := GetLatestPhasesByGameIds(c, []*datastore.Key{gameId})[0]
	var err error
	self.Id, err = datastore.Put(c, datastore.NewKey(c, phaseKind, "", int64(self.Ordinal), gameId), self)
	common.AssertOkError(err)
	if latest == nil || latest.Ordinal <= self.Ordinal {
		common.MemDel(c, latestPhaseByGameIdKey(gameId))
	}
	return self
}

func findLatestPhaseByGameId(c appengine.Context, gameId *datastore.Key) *Phase {
	var phases []Phase
	ids, err := datastore.NewQuery(phaseKind).Ancestor(gameId).Order("-Ordinal").Limit(1).GetAll(c, &phases)
	common.AssertOkError(err)
	for index, id := range ids {
		phases[index].Id = id
	}
	if len(phases) == 0 {
		return nil
	}
	return &phases[0]
}

func GetLatestPhasesByGameIds(c appengine.Context, gameIds []*datastore.Key) (result Phases) {
	cacheKeys := make([]string, len(gameIds))
	values := make([]interface{}, len(gameIds))
	funcs := make([]func() interface{}, len(gameIds))
	for index, id := range gameIds {
		var phase Phase
		values[index] = &phase
		cacheKeys[index] = latestPhaseByGameIdKey(id)
		idCopy := id
		funcs[index] = func() interface{} {
			return findLatestPhaseByGameId(c, idCopy)
		}
	}
	existed := common.MemoizeMulti(c, cacheKeys, values, funcs)
	result = make(Phases, len(gameIds))
	for index, value := range values {
		if existed[index] {
			result[index] = value.(*Phase)
		}
	}
	return
}
