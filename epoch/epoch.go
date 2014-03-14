package epoch

import (
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/kcwraps/kol"
)

const (
	epochKey = "github.com/zond/diplicity/epoch.Epoch"
)

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
	result = epoch.At
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
			if err = Set(c.DB(), currently); err != nil {
				panic(err)
			}
			c.Debugf("Epoch %v", currently)
		}
	}()
	return
}
