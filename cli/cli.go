package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/web"
	"github.com/zond/wsubs/gosubs"
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

	gosubs.Secret = *secret
	token := &gosubs.Token{
		Principal: *email,
		Timeout:   time.Now().Add(time.Second * 10),
	}
	if err := token.Encode(); err != nil {
		panic(err)
	}
	url := fmt.Sprintf("ws://%v:%v/ws?token=%v", *host, *port, token.Encoded)

	ws, err := websocket.Dial(url, "tcp", "http://localhost/")
	if err != nil {
		panic(err)
	}

	if *join != "" {
		base64DecodedId, err := base64.URLEncoding.DecodeString(*join)
		if err != nil {
			panic(err)
		}
		if err := websocket.JSON.Send(ws, gosubs.Message{
			Type: common.UpdateType,
			Object: &gosubs.Object{
				URI: fmt.Sprintf("/games/%v", *join),
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
