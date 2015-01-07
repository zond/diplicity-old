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

func getDB(db *unbolted.DB) (result time.Duration, err error) {
	if err = db.View(func(tx *unbolted.TX) (err error) {
		result, err = Get(tx)
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

func set(d *unbolted.DB, at time.Duration) (err error) {
	epoch := &Epoch{
		Id: unbolted.Id(epochKey),
		At: at,
	}
	err = d.Set(epoch)
	return
}

func Start(c srv.SkinnyContext) (err error) {
	startedAt, err := getDB(c.DB())
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
			if err = set(c.DB(), currently); err != nil {
				panic(err)
			}
			c.Debugf("Epoch %v", currently)
		}
	}()
	return
}
