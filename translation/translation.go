package translation

import (
	"common"
	"net/http"
	"strings"
)

var en = map[string]string{
	"Diplicity": "Diplicity",
	"Menu":      "Menu",
	"England":   "England",
	"Spring":    "Spring",
	"Movement":  "Movement",
	"Games":     "Games",
	"Join":      "Join",
	"Create":    "Create",
	"History":   "History",
	"Home":      "Home",
	"Map":       "Map",
	"Chat":      "Chat",
	"Orders":    "Orders",
	"Results":   "Results",
}

var languages = map[string]map[string]string{
	"en": en,
}

func GetLanguage(r *http.Request) string {
	bestLanguage := common.MostAccepted(r, "default", "Accept-Language")
	parts := strings.Split(bestLanguage, "-")
	return parts[0]
}

func GetTranslations(r *http.Request) (result map[string]string) {
	language := GetLanguage(r)
	result, ok := languages[language]
	if !ok {
		result = en
	}
	return
}
