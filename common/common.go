package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	cla "github.com/zond/godip/classical/common"
	"github.com/zond/godip/classical/start"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/graph"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type Mailer interface {
	SendMail(from, subject, message string, recips ...string) error
	MailAddress() string
}

type SkinnyContext interface {
	gosubs.Logger
	gosubs.SubscriptionManager
	Mailer
	DB() *kol.DB
	BetweenTransactions(func(c SkinnyContext))
	Transact(func(c SkinnyContext) error) error
	Env() string
}

type skinnyWeb struct {
	*Web
	db *kol.DB
}

func (self skinnyWeb) BetweenTransactions(f func(c SkinnyContext)) {
	self.db.BetweenTransactions(func(d *kol.DB) {
		self.db = d
		f(self)
	})
}

func (self skinnyWeb) Transact(f func(c SkinnyContext) error) error {
	return self.db.Transact(func(d *kol.DB) error {
		self.db = d
		return f(self)
	})
}

func (self skinnyWeb) DB() *kol.DB {
	return self.db
}

type skinnyWSContext struct {
	WSContext
}

func (self skinnyWSContext) BetweenTransactions(f func(c SkinnyContext)) {
	self.WSContext.BetweenTransactions(func(c WSContext) {
		f(skinnyWSContext{c})
	})
}

func (self skinnyWSContext) Transact(f func(c SkinnyContext) error) error {
	return self.WSContext.Transact(func(c WSContext) error {
		return f(skinnyWSContext{c})
	})
}

func GetLanguage(r *http.Request) string {
	bestLanguage := MostAccepted(r, "default", "Accept-Language")
	parts := strings.Split(bestLanguage, "-")
	return parts[0]
}

type ChatFlag int

const (
	ChatPrivate = 1 << iota
	ChatGroup
	ChatConference
)

type ChatChannel map[dip.Nation]bool

func (self ChatChannel) Clone() (result ChatChannel) {
	result = ChatChannel{}
	for nation, _ := range self {
		result[nation] = true
	}
	return
}

type ChatFlagOption struct {
	Id          ChatFlag
	Name        string
	Translation string
}

var ChatFlagOptions = []ChatFlagOption{
	ChatFlagOption{
		Id:   ChatPrivate,
		Name: "Private press",
	},
	ChatFlagOption{
		Id:   ChatGroup,
		Name: "Group press",
	},
	ChatFlagOption{
		Id:   ChatConference,
		Name: "Conference press",
	},
}

const (
	regular = iota
	nilCache
)

const (
	ClassicalString                   = "classical"
	RandomString                      = "random"
	PreferencesString                 = "preferences"
	BeforeGamePhaseType dip.PhaseType = "BeforeGame"
	AfterGamePhaseType  dip.PhaseType = "AfterGame"
	Anonymous           dip.Nation    = "Anonymous"
)

type GameState int

const (
	GameStateCreated GameState = iota
	GameStateStarted
	GameStateEnded
)

type SecretFlag int

const (
	SecretBeforeGame = 1 << iota
	SecretDuringGame
	SecretAfterGame
)

type AllocationMethod struct {
	Id          string
	Name        string
	Translation string
}

type AllocationMethodSlice []AllocationMethod

func (self AllocationMethodSlice) Len() int {
	return len(self)
}

func (self AllocationMethodSlice) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}

func (self AllocationMethodSlice) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

var randomAllocationMethod = AllocationMethod{
	Id:   RandomString,
	Name: "Random",
}

var preferencesAllocationMethod = AllocationMethod{
	Id:   PreferencesString,
	Name: "Preferences",
}

var AllocationMethods = AllocationMethodSlice{
	randomAllocationMethod,
	preferencesAllocationMethod,
}

var AllocationMethodMap = map[string]AllocationMethod{
	RandomString:      randomAllocationMethod,
	PreferencesString: preferencesAllocationMethod,
}

type Variant struct {
	Id          string
	Name        string
	Translation string
	PhaseTypes  []dip.PhaseType
	Nations     []dip.Nation
	Colors      map[dip.Nation]string
	Graph       *graph.Graph
}

func (self Variant) JSONNations() string {
	b, _ := json.Marshal(self.Nations)
	return string(b)
}

type VariantSlice []Variant

func (self VariantSlice) Len() int {
	return len(self)
}

func (self VariantSlice) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}

func (self VariantSlice) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

var classicalVariant = Variant{
	Id:         ClassicalString,
	Name:       "Classical",
	PhaseTypes: cla.PhaseTypes,
	Nations:    cla.Nations,
	Colors: map[dip.Nation]string{
		cla.Austria: "#afe773",
		cla.England: "#483c6c",
		cla.France:  "#5693aa",
		cla.Germany: "#ff8b66",
		cla.Italy:   "#1b6c61",
		cla.Russia:  "#8d5e68",
		cla.Turkey:  "#ffdb66",
	},
	Graph: start.Graph(),
}

var Variants = VariantSlice{
	classicalVariant,
}

var VariantMap = map[string]Variant{
	ClassicalString: classicalVariant,
}

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score, _ = strconv.ParseFloat(match[3], 64)
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}

func HostURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v/reload", scheme, r.Host)
}
