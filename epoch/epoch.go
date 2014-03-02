package epoch

import (
	"time"

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
