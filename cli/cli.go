package main

import (
	"encoding/json"
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

type cli struct {
	host   string
	port   int
	secret string
}

func (self *cli) token(email string) (result string, err error) {
	gosubs.Secret = self.secret
	token := &gosubs.Token{
		Principal: email,
		Timeout:   time.Now().Add(time.Second * 10),
	}
	if err = token.Encode(); err != nil {
		return
	}
	result = token.Encoded
	return
}

func (self *cli) connect(email string) (ws *websocket.Conn, receiver chan gosubs.Message, err error) {
	token, err := self.token(email)
	if ws, err = websocket.Dial(fmt.Sprintf("ws://%v:%v/ws?token=%v", self.host, self.port, token), "tcp", "http://localhost/"); err != nil {
		return
	}
	receiver = make(chan gosubs.Message, 1024)
	go func() {
		var err error
		for err == nil {
			mess := gosubs.Message{}
			if err = websocket.JSON.Receive(ws, &mess); err == nil {
				receiver <- mess
			}
		}
	}()
	return
}

func (self *cli) send(email string, mess gosubs.Message) (err error) {
	ws, _, err := self.connect(email)
	if err != nil {
		return
	}
	defer ws.Close()
	if err = websocket.JSON.Send(ws, mess); err != nil {
		return
	}
	return
}

func (self *cli) req(path string) (result io.ReadCloser, err error) {
	token, err := self.token(web.Admin)
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf("http://%v:%v%v?token=%v", self.host, self.port, path, token))
	if err != nil {
		panic(err)
	}
	result = resp.Body
	return
}

func (self *cli) get(path string) (result interface{}, err error) {
	bod, err := self.req(path)
	err = json.NewDecoder(bod).Decode(&result)
	return
}

func (self *cli) game(id string) (result web.AdminGameState, err error) {
	bod, err := self.req("/admin/games/" + id)
	err = json.NewDecoder(bod).Decode(&result)
	return
}

func (self *cli) rpc(email string, method string, data interface{}) (result interface{}, err error) {
	ws, rec, err := self.connect(email)
	if err != nil {
		return
	}
	defer ws.Close()
	id := fmt.Sprint(rand.Int63())
	if err = websocket.JSON.Send(ws, gosubs.Message{
		Type: gosubs.RPCType,
		Method: &gosubs.Method{
			Name: "Commit",
			Id:   id,
			Data: data,
		},
	}); err != nil {
		return
	}
	var mess gosubs.Message
	for mess = <-rec; mess.Type != gosubs.RPCType || mess.Method.Id != id; mess = <-rec {
	}
	result = mess.Method.Data
	return
}

func (self *cli) commit(email string, phaseId interface{}) (err error) {
	_, err = self.rpc(email, "Commit", map[string]interface{}{
		"PhaseId": phaseId,
	})
	return
}

func main() {
	host := flag.String("host", "localhost", "The host to connect to.")
	port := flag.Int("port", 8080, "The port to connect to.")
	email := flag.String("email", "", "The email to fake authenticating as. Mandatory unless url is provided.")
	secret := flag.String("secret", gosubs.Secret, "The token secret of the server.")
	join := flag.String("join", "", "A game to join as the provided email.")
	url := flag.String("url", "", "A url to fetch authenticated as Admin.")
	commit := flag.String("commit", "", "A game to commit the latest phase as the provided email.")
	commitAll := flag.String("commit_all", "", "A game to commit the latest phase as all members.")

	flag.Parse()

	cli := &cli{
		host:   *host,
		port:   *port,
		secret: *secret,
	}

	if *url != "" {
		bod, err := cli.req(*url)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, bod)
	} else {
		if *join != "" {
			if *email == "" {
				flag.Usage()
				return
			}

			g, err := cli.game(*join)
			if err != nil {
				panic(err)
			}
			if err := cli.send(*email, gosubs.Message{
				Type: gosubs.UpdateType,
				Object: &gosubs.Object{
					URI: fmt.Sprintf("/games/%v", *join),
					Data: game.GameState{
						Game: &game.Game{
							Id: g.Game.Id,
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

		if *commitAll != "" {
			g, err := cli.game(*commitAll)
			if err != nil {
				panic(err)
			}
			for _, memb := range g.Members {
				if err := cli.commit(memb.User.Email, g.Phases[0].Id); err != nil {
					panic(err)
				}
			}
		}

		if *commit != "" {
			if *email == "" {
				flag.Usage()
				return
			}

			g, err := cli.game(*commit)
			if err != nil {
				panic(err)
			}
			if err := cli.commit(*email, g.Phases[0].Id); err != nil {
				panic(err)
			}
		}
	}

}
