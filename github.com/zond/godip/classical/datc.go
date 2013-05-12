package classical

import (
	"fmt"
	cla "github.com/zond/godip/classical/common"
	"github.com/zond/godip/classical/orders"
	"github.com/zond/godip/classical/start"
	dip "github.com/zond/godip/common"
	"regexp"
	"strings"
	"time"
)

func init() {
	for _, prov := range start.Graph().Provinces() {
		datcProvinces[string(prov)] = prov
	}
}

var datcPhaseTypes = map[string]dip.PhaseType{
	"movement":   cla.Movement,
	"adjustment": cla.Adjustment,
	"retreat":    cla.Retreat,
}

var datcSeasons = map[string]dip.Season{
	"spring": cla.Spring,
	"fall":   cla.Fall,
}

var datcNationalities = map[string]dip.Nation{
	"england": cla.England,
	"france":  cla.France,
	"germany": cla.Germany,
	"russia":  cla.Russia,
	"austria": cla.Austria,
	"italy":   cla.Italy,
	"turkey":  cla.Turkey,
	"germnay": cla.Germany,
}

var datcUnitTypes = map[string]dip.UnitType{
	"a": cla.Army,
	"f": cla.Fleet,
}

var datcProvinces = map[string]dip.Province{}

func DATCPhase(season string, year int, typ string) dip.Phase {
	phaseType, ok := datcPhaseTypes[strings.ToLower(typ)]
	if !ok {
		panic(fmt.Errorf("Unknown phase type %#v", typ))
	}
	phaseSeason, ok := datcSeasons[strings.ToLower(season)]
	if !ok {
		panic(fmt.Errorf("Unknown season %#v", season))
	}
	return &phase{
		season: phaseSeason,
		typ:    phaseType,
		year:   year,
	}
}

func DATCProvince(n string) (result dip.Province) {
	var ok bool
	result, ok = datcProvinces[strings.ToLower(n)]
	if !ok {
		panic(fmt.Errorf("Unknown province %#v", n))
	}
	return
}

var datcOrderTypes = map[*regexp.Regexp]func([]string) (dip.Province, dip.Adjudicator){
	regexp.MustCompile("(?i)^(A|F)\\s+(\\S+)\\s*-\\s*(\\S+)(\\s+via\\s+convoy)?$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		if m[4] == "" {
			order = orders.Move(DATCProvince(m[2]), DATCProvince(m[3]))
		} else {
			order = orders.Move(DATCProvince(m[2]), DATCProvince(m[3])).ViaConvoy()
		}
		return
	},
	regexp.MustCompile("^(?i)remove\\s+((A|F)\\s+)?(\\S+)$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[3])
		order = orders.Disband(prov, time.Now())
		return
	},
	regexp.MustCompile("^(?i)(A|F)\\s+(\\S+)\\s+disband$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Disband(prov, time.Now())
		return
	},
	regexp.MustCompile("^(?i)(A|F)\\s+(\\S+)\\s+S(UPP\\S*)?\\s+(A|F)\\s+([^-\\s]+)$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Support(DATCProvince(m[2]), DATCProvince(m[5]))
		return
	},
	regexp.MustCompile("^(?i)(A|F)\\s+(\\S+)\\s+c(onv\\S*)?\\s+(A|F)\\s+(\\S+)\\s*-\\s*(\\S+)$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Convoy(DATCProvince(m[2]), DATCProvince(m[5]), DATCProvince(m[6]))
		return
	},
	regexp.MustCompile("^(?i)(A|F)\\s+(\\S+)\\s+S(UPP\\S*)?\\s+((A|F)\\s+)?(\\S+)\\s*-\\s*(\\S+)$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Support(DATCProvince(m[2]), DATCProvince(m[6]), DATCProvince(m[7]))
		return
	},
	regexp.MustCompile("^(?i)(A|F)\\s+(\\S+)\\s+H(OLD)?$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Hold(DATCProvince(m[2]))
		return
	},
	regexp.MustCompile("^(?i)build\\s+(A|F)\\s+(\\S+)\\s*$"): func(m []string) (prov dip.Province, order dip.Adjudicator) {
		prov = DATCProvince(m[2])
		order = orders.Build(DATCProvince(m[2]), DATCUnitType(m[1]), time.Now())
		return
	},
}

func DATCOrder(text string) (province dip.Province, order dip.Adjudicator) {
	var match []string
	for reg, gen := range datcOrderTypes {
		if match = reg.FindStringSubmatch(text); match != nil {
			return gen(match)
		}
	}
	panic(fmt.Errorf("Unknown order text: %#v", text))
}

func DATCNation(typ string) (result dip.Nation) {
	var ok bool
	result, ok = datcNationalities[strings.ToLower(typ)]
	if !ok {
		panic(fmt.Errorf("Unknown nationality: %#v", typ))
	}
	return
}

func DATCUnitType(typ string) (result dip.UnitType) {
	var ok bool
	result, ok = datcUnitTypes[strings.ToLower(typ)]
	if !ok {
		panic(fmt.Errorf("Unknown unit type: %#v", typ))
	}
	return
}
