package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	cla "github.com/zond/godip/classical/common"
	"github.com/zond/godip/classical/start"
	dip "github.com/zond/godip/common"
	"github.com/zond/godip/graph"
)

func RenderJSON(w http.ResponseWriter, i interface{}) {
	SetContentType(w, "application/json; charset=UTF-8", false)
	MustEncodeJSON(w, i)
}

func SetContentType(w http.ResponseWriter, t string, cache bool) {
	w.Header().Set("Content-Type", t)
	w.Header().Set("Vary", "Accept")
	if cache {
		w.Header().Set("Cache-Control", "public, max-age=864000")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
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
	return string(MustMarshalJSON(self.Nations))
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

func Prettify(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt64(s string) (result int64) {
	var err error
	if result, err = strconv.ParseInt(s, 10, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt(s string) (result int) {
	var err error
	if result, err = strconv.Atoi(s); err != nil {
		panic(err)
	}
	return
}

func MustParseURL(s string) (result *url.URL) {
	var err error
	if result, err = url.Parse(s); err != nil {
		panic(err)
	}
	return
}

func MustMarshalJSON(i interface{}) (result []byte) {
	var err error
	result, err = json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return
}

func MustUnmarshalJSON(data []byte, result interface{}) {
	if err := json.Unmarshal(data, result); err != nil {
		panic(err)
	}
}

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
		panic(err)
	}
}

func MustDecodeJSON(r io.Reader, result interface{}) {
	if err := json.NewDecoder(r).Decode(result); err != nil {
		panic(err)
	}
}

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score = MustParseFloat64(match[3])
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
