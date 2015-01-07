package srv

import (
	"github.com/zond/unbolted"
	"github.com/zond/unbolted/pack"
	"github.com/zond/wsubs/gosubs"
)

type WSContext interface {
	pack.SubContext
	AfterTransaction(func(WSContext) error) error
	View(func(WSTXContext) error) error
	Update(func(WSTXContext) error) error
	Mailer
	Env() string
	Diet() SkinnyContext
	Secret() string
}

type WSTXContext interface {
	WSContext
	TX() *unbolted.TX
	TXDiet() SkinnyTXContext
}

type SkinnyContext interface {
	gosubs.Logger
	gosubs.SubscriptionManager
	DB() *unbolted.DB
	AfterTransaction(func(SkinnyContext) error) error
	View(func(SkinnyTXContext) error) error
	Update(func(SkinnyTXContext) error) error
	Mailer
	Env() string
	Secret() string
}

type SkinnyTXContext interface {
	SkinnyContext
	TX() *unbolted.TX
}
