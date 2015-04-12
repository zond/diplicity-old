package epoch

import (
	"time"

	"sync/atomic"

	"github.com/zond/diplicity/srv"
	"github.com/zond/unbolted"
)

const (
	epochKey = "github.com/zond/diplicity/epoch.Epoch"
)

var deltaPoint int64 = time.Now().UnixNano()

type Epoch struct {
	Id unbolted.Id
	At time.Duration
}

func getDB(c srv.Context) (result time.Duration, err error) {
	if err = c.View(func(c srv.Context) (err error) {
		result, err = Get(c.TX())
		return
	}); err != nil {
		return
	}
	return
}

func Get(tx *unbolted.TX) (result time.Duration, err error) {
	epoch := &Epoch{
		Id: unbolted.Id(epochKey),
	}
	if err = tx.Get(epoch); err != nil {
		if err == unbolted.ErrNotFound {
			err = nil
		} else {
			return
		}
	}
	result = epoch.At + time.Now().Sub(time.Unix(0, atomic.LoadInt64(&deltaPoint)))
	return
}

func set(c srv.Context, at time.Duration) (err error) {
	epoch := &Epoch{
		Id: unbolted.Id(epochKey),
		At: at,
	}
	return c.Update(func(c srv.Context) error { return c.TX().Set(epoch) })
}

func Start(c srv.Context) (err error) {
	startedAt, err := getDB(c)
	if err != nil {
		return
	}
	c.Infof("Started at epoch %v", startedAt)
	startedTime := time.Now()
	var currently time.Duration
	go func() {
		for {
			time.Sleep(time.Minute)
			currently = time.Now().Sub(startedTime) + startedAt
			atomic.StoreInt64(&deltaPoint, int64(time.Now().UnixNano()))
			if err = set(c, currently); err != nil {
				panic(err)
			}
			c.Debugf("Epoch %v", currently)
		}
	}()
	return
}
