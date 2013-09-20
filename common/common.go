package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	cla "github.com/zond/godip/classical/common"
	dip "github.com/zond/godip/common"
	"github.com/zond/kcwraps/kol"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Subscriber func(i interface{}, op string)

type Subscription struct {
	Name       string
	Subscriber Subscriber
	Object     interface{}
	Query      *kol.Query
}

type Logger interface {
	Fatalf(format string, params ...interface{})
	Errorf(format string, params ...interface{})
	Infof(format string, params ...interface{})
	Debugf(format string, params ...interface{})
	Tracef(format string, params ...interface{})
}

type Context interface {
	Logger
	DB() *kol.DB
}

type JSON struct {
	Data interface{}
}

func (self JSON) Get(key string) JSON {
	return JSON{self.Data.(map[string]interface{})[key]}
}

func (self JSON) Overwrite(dest interface{}) {
	MustUnmarshalJSON(MustMarshalJSON(self.Data), dest)
}

func (self JSON) GetString(key string) string {
	return self.Data.(map[string]interface{})[key].(string)
}

func GetLanguage(r *http.Request) string {
	bestLanguage := MostAccepted(r, "default", "Accept-Language")
	parts := strings.Split(bestLanguage, "-")
	return parts[0]
}

type ChatFlag int

const (
	ChatWhite ChatFlag = 1 << iota
	ChatGrey
	ChatBlack
	ChatPrivate
	ChatGroup
	ChatConference
)

type ChatFlagOption struct {
	Id          ChatFlag
	Name        string
	Translation string
}

var ChatFlagOptions = []ChatFlagOption{
	ChatFlagOption{
		Id:   ChatWhite,
		Name: "White press",
	},
	ChatFlagOption{
		Id:   ChatGrey,
		Name: "Grey press",
	},
	ChatFlagOption{
		Id:   ChatBlack,
		Name: "Black press",
	},
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
	StandardString    = "standard"
	RandomString      = "random"
	PreferencesString = "preferences"
)

type AllocationMethod struct {
	Id          string
	Name        string
	Translation string
}

type AllocationMethods []AllocationMethod

func (self AllocationMethods) Len() int {
	return len(self)
}

func (self AllocationMethods) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}

func (self AllocationMethods) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

var AllocationMethodMap = map[string]AllocationMethod{
	RandomString: AllocationMethod{
		Id:   RandomString,
		Name: "Random",
	},
	PreferencesString: AllocationMethod{
		Id:   PreferencesString,
		Name: "Preferences",
	},
}

type Variant struct {
	Id          string
	Name        string
	Translation string
	PhaseTypes  []dip.PhaseType
	Nations     []dip.Nation
}

func (self Variant) JSONNations() string {
	return string(MustMarshalJSON(self.Nations))
}

type Variants []Variant

func (self Variants) Len() int {
	return len(self)
}

func (self Variants) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}

func (self Variants) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

var VariantMap = map[string]Variant{
	StandardString: Variant{
		Id:         StandardString,
		Name:       "Standard",
		PhaseTypes: cla.PhaseTypes,
		Nations:    cla.Nations,
	},
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
