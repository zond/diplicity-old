package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
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
	email := flag.String("email", "", "The email to fake authenticating as. Mandatory unless url is provided.")
	secret := flag.String("secret", gosubs.Secret, "The token secret of the server.")
	join := flag.String("join", "", "A game to join as the provided email.")
	url := flag.String("url", "", "A url to fetch authenticated as Admin.")
	commit := flag.String("commit", "", "A game to commit the latest phase as the provided email.")

	flag.Parse()

	if *url != "" {
		*email = web.Admin
	}

	gosubs.Secret = *secret
	token := &gosubs.Token{
		Principal: *email,
		Timeout:   time.Now().Add(time.Second * 10),
	}
	if err := token.Encode(); err != nil {
		panic(err)
	}

	if *url != "" {
		cli := &http.Client{}
		resp, err := cli.Get(fmt.Sprintf("http://%v:%v/%v?token=%v", *host, *port, *url, token.Encoded))
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, resp.Body)
	} else {
		if *email == "" {
			flag.Usage()
			return
		}

		ws, err := websocket.Dial(fmt.Sprintf("ws://%v:%v/ws?token=%v", *host, *port, token.Encoded), "tcp", "http://localhost/")
		if err != nil {
			panic(err)
		}
		receiver := make(chan gosubs.Message, 1024)
		go func() {
			var err error
			for err == nil {
				mess := gosubs.Message{}
				if err = websocket.JSON.Receive(ws, &mess); err == nil {
					receiver <- mess
				}
			}
		}()

		if *join != "" {
			base64DecodedId, err := base64.URLEncoding.DecodeString(*join)
			if err != nil {
				panic(err)
			}
			if err := websocket.JSON.Send(ws, gosubs.Message{
				Type: gosubs.UpdateType,
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

		if *commit != "" {
			if err := websocket.JSON.Send(ws, gosubs.Message{
				Type: gosubs.SubscribeType,
				Object: &gosubs.Object{
					URI: fmt.Sprintf("/games/%v", *commit),
				},
			}); err != nil {
				panic(err)
			}
			mess := <-receiver
			id := fmt.Sprint(rand.Int63())
			if err := websocket.JSON.Send(ws, gosubs.Message{
				Type: gosubs.RPCType,
				Method: &gosubs.Method{
					Name: "Commit",
					Id:   id,
					Data: map[string]interface{}{
						"PhaseId": mess.Object.Data.(map[string]interface{})["Phase"].(map[string]interface{})["Id"],
					},
				},
			}); err != nil {
				panic(err)
			}
			for mess = <-receiver; mess.Type != gosubs.RPCType || mess.Method.Id != id; mess = <-receiver {
			}
		}
		ws.Close()
	}

}
