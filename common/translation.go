package common

import (
	"net/http"
	"strings"
)

var en = map[string]string{
	"Log in to see your games": "Log in to see your games",
	"Log in to create a game":  "Log in to create a game",
	"Private":                  "Private",
	"Public":                   "Public",
	"forming":                  "forming",
	"undecided":                "undecided",
	"Diplicity":                "Diplicity",
	"Menu":                     "Menu",
	"Logout":                   "Logout",
	"Login":                    "Login",
	"nations":                  "{'Austria': 'Austria', 'England': 'England', 'France': 'France', 'Germany': 'Germany', 'Italy': 'Italy', 'Russia': 'Russia', 'Turkey': 'Turkey'}",
	"seasons":                  "{'Spring': 'Spring', 'Fall': 'Fall'}",
	"phase_types":              "{'Movement': 'Movement', 'Retreat': 'Retreat', 'Adjustment': 'Adjustment'}",
	"England":                  "England",
	"Variant":                  "Variant",
	"Standard":                 "Standard",
	"Spring":                   "Spring",
	"Movement":                 "Movement",
	"Games":                    "Games",
	"Join":                     "Join",
	"Create":                   "Create",
	"History":                  "History",
	"Home":                     "Home",
	"Map":                      "Map",
	"Chat":                     "Chat",
	"Orders":                   "Orders",
	"Results":                  "Results",
}

var languages = map[string]map[string]string{
	"en": en,
}

func getLanguage(r *http.Request) string {
	bestLanguage := MostAccepted(r, "default", "Accept-Language")
	parts := strings.Split(bestLanguage, "-")
	return parts[0]
}

func getTranslations(r *http.Request) (result map[string]string) {
	language := getLanguage(r)
	result, ok := languages[language]
	if !ok {
		result = en
	}
	return
}
