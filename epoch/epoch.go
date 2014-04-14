package epoch

import (
	"time"

	"sync/atomic"

	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
)

const (
	epochKey = "github.com/zond/diplicity/epoch.Epoch"
)

var deltaPoint int64 = time.Now().UnixNano()

type Epoch struct {
	Id kol.Id
	At time.Duration
}

func Get(d *kol.DB) (result time.Duration, err error) {
	epoch := &Epoch{
		Id: kol.Id(epochKey),
	}
	if err = d.Get(epoch); err != nil {
		if err == kol.NotFound {
			err = nil
		} else {
			return
		}
	}
	result = epoch.At + time.Now().Sub(time.Unix(0, atomic.LoadInt64(&deltaPoint)))
	return
}

func Set(d *kol.DB, at time.Duration) (err error) {
	epoch := &Epoch{
		Id: kol.Id(epochKey),
		At: at,
	}
	err = d.Set(epoch)
	return
}

func Start(c common.SkinnyContext) (err error) {
	startedAt, err := Get(c.DB())
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
			if err = Set(c.DB(), currently); err != nil {
				panic(err)
			}
			c.Debugf("Epoch %v", currently)
		}
	}()
	return
}
