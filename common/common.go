package common

import (
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	cla "github.com/zond/godip/classical/common"
	claOrders "github.com/zond/godip/classical/orders"
	"github.com/zond/godip/classical/start"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/graph"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

type Translator interface {
	I(phrase string, args ...interface{}) (result string, err error)
}

const (
	DefaultSecret       = "something very very secret"
	SubscriptionTimeout = time.Minute * 15
)

type Mailer interface {
	SendMail(fromName, replyTo, subject, message string, recips []string) error
	ReceiveAddress() string
	SendAddress() string
}

type SkinnyContext interface {
	gosubs.Logger
	gosubs.SubscriptionManager
	Mailer
	DB() *kol.DB
	BetweenTransactions(func(c SkinnyContext))
	Transact(func(c SkinnyContext) error) error
	Env() string
	Secret() string
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

type Consequence int

const (
	ReliabilityHit Consequence = 1 << iota
	NoWait
	Surrender
)

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

type ConsequenceOption struct {
	Id          Consequence
	Name        string
	Translation string
}

var ConsequenceOptions = []ConsequenceOption{
	ConsequenceOption{
		Id:   ReliabilityHit,
		Name: "Reliability hit",
	},
	ConsequenceOption{
		Id:   NoWait,
		Name: "No wait",
	},
	ConsequenceOption{
		Id:   Surrender,
		Name: "Surrender",
	},
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

type EndReason string

const (
	ClassicalString                   = "classical"
	RandomString                      = "random"
	PreferencesString                 = "preferences"
	BeforeGamePhaseType dip.PhaseType = "BeforeGame"
	AfterGamePhaseType  dip.PhaseType = "AfterGame"
	Anonymous           dip.Nation    = "Anonymous"
	ZeroActiveMembers   EndReason     = "ZeroActiveMembers"
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
	Id                   string
	Name                 string
	Translation          string
	PhaseTypes           []dip.PhaseType
	Nations              []dip.Nation
	Colors               map[dip.Nation]string
	Graph                *graph.Graph
	OrderTypes           []dip.OrderType
	UnitTypes            []dip.UnitType
	NationAbbrevs        map[string]dip.Nation
	OrderTypeAbbrevs     map[string]dip.OrderType
	UnitTypeAbbrevs      map[string]dip.UnitType
	SupplyCenters        map[dip.Province]dip.Nation
	SelectableProvinces  []dip.Province
	ColorizableProvinces []dip.Province
	Seasons              []dip.Season
}

func (self Variant) JSONNations() string {
	b, _ := json.Marshal(self.Nations)
	return string(b)
}

type VariantSlice []*Variant

func (self VariantSlice) Len() int {
	return len(self)
}

func (self VariantSlice) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}

func (self VariantSlice) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

var classicalVariant = &Variant{
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
	OrderTypes:       claOrders.OrderTypes(),
	UnitTypes:        cla.UnitTypes,
	Graph:            start.Graph(),
	SupplyCenters:    start.SCs(),
	OrderTypeAbbrevs: map[string]dip.OrderType{},
	NationAbbrevs:    map[string]dip.Nation{},
	UnitTypeAbbrevs:  map[string]dip.UnitType{},
	Seasons:          cla.Seasons,
}

func init() {
	for _, orderType := range classicalVariant.OrderTypes {
		i := 1
		for {
			if _, found := classicalVariant.OrderTypeAbbrevs[string(orderType)[0:i]]; !found {
				break
			}
			i++
		}
		classicalVariant.OrderTypeAbbrevs[string(orderType)[0:i]] = orderType
	}
	for _, unitType := range classicalVariant.UnitTypes {
		i := 1
		for {
			if _, found := classicalVariant.UnitTypeAbbrevs[string(unitType)[0:i]]; !found {
				break
			}
			i++
		}
		classicalVariant.UnitTypeAbbrevs[string(unitType)[0:i]] = unitType
	}
	for _, nation := range classicalVariant.Nations {
		i := 1
		for {
			if _, found := classicalVariant.NationAbbrevs[string(nation)[0:i]]; !found {
				break
			}
			i++
		}
		classicalVariant.NationAbbrevs[string(nation)[0:i]] = nation
	}
	/*
		All provinces that are either coastless or are the coasts themselves
	*/
	hasCoasts := map[dip.Province]bool{}
	all := map[dip.Province]bool{}
	for _, prov := range classicalVariant.Graph.Provinces() {
		all[prov] = true
		if prov != prov.Super() {
			hasCoasts[prov.Super()] = true
		}
	}
	for prov, _ := range all {
		if prov != prov.Super() || !hasCoasts[prov] {
			classicalVariant.SelectableProvinces = append(classicalVariant.SelectableProvinces, prov)
		}
	}
	/*
		All provinces that arent coasts
	*/
	supers := map[dip.Province]bool{}
	for _, prov := range classicalVariant.Graph.Provinces() {
		supers[prov.Super()] = true
	}
	for prov, _ := range supers {
		classicalVariant.ColorizableProvinces = append(classicalVariant.ColorizableProvinces, prov)
	}
}

var Variants = VariantSlice{
	classicalVariant,
}

var VariantMap = map[string]*Variant{
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

const (
	UnsubscribeMessageEmail = iota
	UnsubscribePhaseEmail
)

const (
	EmailTemplate = `%v
----
%v
%v`
)

type UnsubscribeTag struct {
	T int
	U kol.Id
	H []byte
}

func (self *UnsubscribeTag) Hash(secret string) []byte {
	h := sha1.New()
	h.Write(self.U)
	h.Write([]byte(secret))
	return h.Sum(nil)
}

func (self *UnsubscribeTag) Encode() (result string, err error) {
	buf := &bytes.Buffer{}
	baseEnc := base64.NewEncoder(base64.URLEncoding, buf)
	gobEnc := gob.NewEncoder(baseEnc)
	if err = gobEnc.Encode(self); err != nil {
		return
	}
	if err = baseEnc.Close(); err != nil {
		return
	}
	result = buf.String()
	return
}

func DecodeUnsubscribeTag(secret string, s string) (result *UnsubscribeTag, err error) {
	buf := bytes.NewBufferString(s)
	dec := gob.NewDecoder(base64.NewDecoder(base64.URLEncoding, buf))
	tag := &UnsubscribeTag{}
	if err = dec.Decode(tag); err != nil {
		return
	}
	wanted := tag.Hash(secret)
	if len(wanted) != len(tag.H) || subtle.ConstantTimeCompare(wanted, tag.H) != 1 {
		err = fmt.Errorf("%+v has wrong hash, wanted %v", tag, wanted)
		return
	}
	result = tag
	return
}
