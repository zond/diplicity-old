package common

import (
	"github.com/zond/kcwraps/kol"
)

var DB *kol.DB

func init() {
	var err error
	if DB, err = kol.New("diplicity"); err != nil {
		panic(err)
	}
}
