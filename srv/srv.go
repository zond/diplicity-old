package srv

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
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
