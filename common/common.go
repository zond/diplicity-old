package common

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

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

func SetContentType(w http.ResponseWriter, typ string) {
	w.Header().Set("Content-Type", typ)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
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
