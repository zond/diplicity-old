package common

import (
	"net/http"
	"strings"
)

var en = map[string]string{
	"White press":       "White press",
	"Grey press":        "Grey press",
	"Black press":       "Black press",
	"Private press":     "Private press",
	"Group press":       "Group press",
	"Leave":             "Leave",
	"Conference press":  "Conference press",
	"5 minutes":         "5 minutes",
	"10 minutes":        "10 minutes",
	"20 minutes":        "20 minutes",
	"30 minutes":        "30 minutes",
	"1 hour":            "1 hour",
	"2 hours":           "2 hours",
	"4 hours":           "4 hours",
	"8 hours":           "8 hours",
	"12 hours":          "12 hours",
	"24 hours":          "24 hours",
	"2 days":            "2 days",
	"3 days":            "3 days",
	"4 days":            "4 days",
	"5 days":            "5 days",
	"1 week":            "1 week",
	"10 days":           "10 days",
	"2 weeks":           "2 weeks",
	"Diplicity Welcome": "Welcome to Diplicity! <br /> What would you like to do?",

	"Log in to see your games": "Log in to see your games",
	"Log in to create a game":  "Log in to create a game",
	"Log in to join a game":    "Log in to join a game",
	"Create a game":            "Create a game",
	"Create a Google account":  "Create your Google account to play",
	"View public games":        "View public games",
	"Learn how to play":        "Learn how to play",
	"Private":                  "Private",
	"Details":                  "Details",
	"View":                     "View",
	"Settings":                 "Settings",
	"Public":                   "Public",
	"Forming":                  "Forming",
	"Undecided":                "Undecided",
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
