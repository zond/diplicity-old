package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/web"
	"github.com/zond/kcwraps/subs"
	"net/http"
	"net/url"
)

func main() {
	host := flag.String("host", "localhost", "The host to connect to.")
	port := flag.Int("port", 8080, "The port to connect to.")
	email := flag.String("email", "", "The email to fake authenticating as. Mandatory.")
	secret := flag.String("secret", web.DefaultSecret, "The cookie store secret of the server.")
	join := flag.String("join", "", "A game to join as the provided email.")

	flag.Parse()

	if *email == "" {
		flag.Usage()
		return
	}

	location, err := url.Parse(fmt.Sprintf("ws://%v:%v/ws", *host, *port))
	if err != nil {
		panic(err)
	}
	origin, err := url.Parse("http://localhost")
	if err != nil {
		panic(err)
	}

	encoded, err := securecookie.EncodeMulti(web.SessionName, map[interface{}]interface{}{
		web.SessionEmail: *email,
	}, securecookie.CodecsFromPairs([]byte(*secret))...)
	if err != nil {
		panic(err)
	}

	cookie := sessions.NewCookie(web.SessionName, encoded, &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 30,
	})

	header := http.Header{
		"Cookie": []string{cookie.String()},
	}

	config := &websocket.Config{
		Location: location,
		Origin:   origin,
		Version:  13,
		Header:   header,
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		return
	}

	if *join != "" {
		base64DecodedId, err := base64.URLEncoding.DecodeString(*join)
		if err != nil {
			panic(err)
		}
		if err := websocket.JSON.Send(ws, subs.Message{
			Type: common.UpdateType,
			Object: &subs.Object{
				URI: "/games/open",
				Data: game.GameState{
					Game: &game.Game{
						Id: base64DecodedId,
					},
					Members: []game.MemberState{
						game.MemberState{
							Member: &game.Member{
								PreferredNations: common.VariantMap[common.ClassicalString].Nations,
							},
						},
					},
				},
			},
		}); err != nil {
			panic(err)
		}
	}

}
