package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/srv"
	"github.com/zond/godip/variants"
	"github.com/zond/unbolted"
	"github.com/zond/wsubs"
)

type cli struct {
	host   string
	port   int
	secret string
}

func (self *cli) token(email string) (result string, err error) {
	token := &wsubs.Token{
		Principal: email,
		Timeout:   time.Now().Add(time.Second * 10),
	}
	if err = token.Encode(self.secret); err != nil {
		return
	}
	result = token.Encoded
	return
}

func (self *cli) connect(email string) (ws *websocket.Conn, receiver chan wsubs.Message, err error) {
	token, err := self.token(email)
	if err != nil {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", self.host, self.port))
	if err != nil {
		return
	}
	netconn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return
	}
	u, err := url.Parse(fmt.Sprintf("ws://%v:%v/ws?token=%v", self.host, self.port, token))
	if err != nil {
		return
	}
	if ws, _, err = websocket.NewClient(netconn, u, nil, 1024, 1024); err != nil {
		return
	}
	receiver = make(chan wsubs.Message, 1024)
	go func() {
		var err error
		for err == nil {
			mess := wsubs.Message{}
			if err = ws.ReadJSON(&mess); err == nil {
				receiver <- mess
			}
		}
	}()
	return
}

func (self *cli) send(email string, mess wsubs.Message) (err error) {
	ws, _, err := self.connect(email)
	if err != nil {
		return
	}
	if err = ws.WriteJSON(mess); err != nil {
		return
	}
	return
}

func (self *cli) post(path string, obj interface{}) (response string, err error) {
	token, err := self.token(srv.Admin)
	cli := &http.Client{}
	buf := &bytes.Buffer{}
	if err = json.NewEncoder(buf).Encode(obj); err != nil {
		return
	}
	resp, err := cli.Post(fmt.Sprintf("http://%v:%v%v?token=%v", self.host, self.port, path, token), "application/json", buf)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	response = string(b)
	return
}

func (self *cli) get(path string) (result io.ReadCloser, err error) {
	token, err := self.token(srv.Admin)
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf("http://%v:%v%v?token=%v", self.host, self.port, path, token))
	if err != nil {
		panic(err)
	}
	result = resp.Body
	return
}

func (self *cli) createUser(email string) (err error) {
	_, err = self.post("/admin/users", map[string]interface{}{
		"Email":         email,
		"Id":            unbolted.Id(email),
		"DiplicityHost": fmt.Sprintf("%v:%v", self.host, self.port),
	})
	return
}

func (self *cli) game(id string) (result game.AdminGameState, err error) {
	bod, err := self.get("/admin/games/" + id)
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
	if err = ws.WriteJSON(wsubs.Message{
		Type: wsubs.RPCType,
		Method: &wsubs.Method{
			Name: "Commit",
			Id:   id,
			Data: data,
		},
	}); err != nil {
		return
	}
	var mess wsubs.Message
	for mess = <-rec; mess.Type != wsubs.RPCType || mess.Method.Id != id; mess = <-rec {
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
	secret := flag.String("secret", srv.DefaultSecret, "The token secret of the server.")
	join := flag.String("join", "", "A game to join as the provided email.")
	path := flag.String("path", "", "A path to fetch from the host authenticated as Admin.")
	commit := flag.String("commit", "", "A game to commit the latest phase as the provided email.")
	commitAll := flag.String("commit_all", "", "A game to commit the latest phase as all members.")
	joinX := flag.Int("join_x", 0, "A number of members to join automatically to the game defined by -join.")
	rollback := flag.String("rollback", "", "A game to rollback.")
	recalc := flag.String("recalc", "", "A game to recalculate options for.")
	until := flag.Int("until", 100000, "A phase ordinal to roll back to. This will be the unresolved phase.")
	reindex := flag.Bool("reindex", false, "Reindex all games in the database.")
	setrank1 := flag.Bool("setrank1", false, "Set rank of all users to 1.")

	flag.Parse()

	cli := &cli{
		host:   *host,
		port:   *port,
		secret: *secret,
	}

	if *path != "" {
		bod, err := cli.get(*path)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, bod)
	} else {
		if *join == "" && *commitAll == "" && *commit == "" && *rollback == "" && *recalc == "" && *reindex == false && *setrank1 == false {
			flag.Usage()
			return
		}

		if *reindex {
			if resp, err := cli.post("/admin/games/reindex", nil); err != nil {
				panic(err)
			} else {
				fmt.Println(resp)
			}
		}

		if *setrank1 {
			if resp, err := cli.post("/admin/users/setrank1", nil); err != nil {
				panic(err)
			} else {
				fmt.Println(resp)
			}
		}

		if *recalc != "" {
			if _, err := cli.post(fmt.Sprintf("/admin/games/%v/recalc", *recalc), map[string]interface{}{}); err != nil {
				panic(err)
			}
		}

		if *rollback != "" {
			if _, err := cli.post(fmt.Sprintf("/admin/games/%v/rollback/%v", *rollback, *until), map[string]interface{}{}); err != nil {
				panic(err)
			}
		}

		if *join != "" {
			if (*email == "" && *joinX < 1) || (*email != "" && *joinX > 0) {
				flag.Usage()
				return
			}

			if *joinX > 0 {
				*email = fmt.Sprintf("%v@dom.tld", *joinX)
			} else if *email != "" {
				*joinX = 1
			}

			g, err := cli.game(*join)
			if err != nil {
				panic(err)
			}
			for ; *joinX > 0; *joinX-- {
				if err := cli.createUser(*email); err != nil {
					panic(err)
				}
				if err := cli.send(*email, wsubs.Message{
					Type: wsubs.UpdateType,
					Object: &wsubs.Object{
						URI: fmt.Sprintf("/games/%v", *join),
						Data: game.GameState{
							Game: &game.Game{
								Id: g.Game.Id,
							},
							Members: []game.MemberState{
								game.MemberState{
									Member: &game.Member{
										PreferredNations: variants.Variants["Classical"].Nations,
									},
								},
							},
						},
					},
				}); err != nil {
					panic(err)
				}
				*email = fmt.Sprintf("%v@dom.tld", *joinX-1)
			}
		}

		if *commitAll != "" {
			g, err := cli.game(*commitAll)
			if err != nil {
				panic(err)
			}
			for _, memb := range g.Members {
				if !memb.NoOrders {
					if err := cli.commit(memb.User.Email, g.Phases[0].Id); err != nil {
						panic(err)
					}
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
