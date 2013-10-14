package web

import (
	"code.google.com/p/go.net/websocket"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/game"
	"github.com/zond/diplicity/openid"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/subs"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var gamePattern = regexp.MustCompile("^/games/(.*)$")

func (self *Web) WS(ws *websocket.Conn) {
	token := ws.Request().URL.Query().Get("token")
	email := ws.Request().URL.Query().Get("email")
	loggedIn := false
	if token != "" {
		timeout := common.MustParseInt64(ws.Request().URL.Query().Get("timeout"))
		if now := time.Now().UnixNano(); timeout < now {
			self.Errorf("\t%v\t%v\t[token too old: %v < %v]", ws.Request().URL, ws.Request().RemoteAddr, timeout, now)
			return
		}
		if correct := self.hash(email, timeout); correct != token {
			self.Errorf("\t%v\t%v\t[bad token: %v != %v]", ws.Request().URL, ws.Request().RemoteAddr, token, correct)
			return
		}
		loggedIn = true
	}

	self.Infof("\t%v\t%v\t%v <-", ws.Request().URL, ws.Request().RemoteAddr, email)

	pack := subs.New(self.db, ws).OnUnsubscribe(func(s *subs.Subscription, reason interface{}) {
		self.Errorf("\t%v\t%v\t%v\t%v\t%v\t[unsubscribing]", ws.Request().URL.Path, ws.Request().RemoteAddr, email, s.URI(), reason)
	}).Logger(func(name string, i interface{}, op string, dur time.Duration) {
		self.Debugf("\t%v\t%v\t%v\t%v\t%v\t%v ->", ws.Request().URL.Path, ws.Request().RemoteAddr, email, op, name, dur)
	})
	defer func() {
		self.Infof("\t%v\t%v\t%v -> [unsubscribing all]", ws.Request().URL.Path, ws.Request().RemoteAddr, email)
		pack.UnsubscribeAll()
	}()

	var start time.Time
	for {
		var message subs.Message
		if err := websocket.JSON.Receive(ws, &message); err == nil {
			start = time.Now()
			func() {
				defer func() {
					if message.Method != nil {
						self.Debugf("\t%v\t%v\t%v\t%v\t%v\t%v <-", ws.Request().URL.Path, ws.Request().RemoteAddr, email, message.Type, message.Method.Name, time.Now().Sub(start))
					}
					if message.Object != nil {
						self.Debugf("\t%v\t%v\t%v\t%v\t%v\t%v <-", ws.Request().URL.Path, ws.Request().RemoteAddr, email, message.Type, message.Object.URI, time.Now().Sub(start))
					}
					if self.logLevel > Trace {
						if message.Method != nil && message.Method.Data != nil {
							self.Tracef("%+v", common.Prettify(message.Method.Data))
						}
						if message.Object != nil && message.Object.Data != nil {
							self.Tracef("%+v", common.Prettify(message.Object.Data))
						}
					}
				}()
				switch message.Type {
				case common.SubscribeType:
					s := pack.New(message.Object.URI)
					switch message.Object.URI {
					case "/games/current":
						if loggedIn {
							self.Errlog(game.SubscribeCurrent(self, s, email))
						}
					case "/games/open":
						if loggedIn {
							self.Errlog(game.SubscribeOpen(self, s, email))
						}
					case "/user":
						if loggedIn {
							self.Errlog(user.SubscribeEmail(self, s, email))
						} else {
							s.Call(&user.User{}, subs.FetchType)
						}
					default:
						if match := gamePattern.FindStringSubmatch(message.Object.URI); match != nil {
							if loggedIn {
								game.SubscribeGame(self, s, match[1], email)
							}
						} else {
							self.Errorf("Unrecognized URI: %+v", message.Object.URI)
						}
					}
				case common.UnsubscribeType:
					pack.Unsubscribe(message.Object.URI)
				case common.CreateType:
					switch message.Object.URI {
					case "/games":
						if loggedIn {
							game.Create(self, subs.JSON{message.Object.Data}, email)
						}
					}
				case common.DeleteType:
					if match := gamePattern.FindStringSubmatch(message.Object.URI); match != nil {
						if loggedIn {
							game.DeleteMember(self, match[1], email)
						}
					} else {
						self.Errorf("Unrecognized URI to delete: %v", message.Object.URI)
					}
				case common.UpdateType:
					if strings.Index(message.Object.URI, "/games/") == 0 {
						if loggedIn {
							game.AddMember(self, subs.JSON{message.Object.Data}, email)
						}
					} else if strings.Index(message.Object.URI, "/user") == 0 {
						if loggedIn {
							user.Update(self, subs.JSON{message.Object.Data}, email)
						}
					}
				case common.RPCType:
					switch message.Method.Name {
					case "GetPossibleSources":
						if loggedIn {
							params := subs.JSON{message.Method.Data}
							if result, err := game.GetPossibleSources(self, params.GetString("GameId"), email); err == nil {
								if err := websocket.JSON.Send(ws, subs.Message{
									Type: common.RPCType,
									Method: &subs.Method{
										Name: message.Method.Name,
										Id:   message.Method.Id,
										Data: result,
									},
								}); err != nil {
									self.Errorf("%v", err)
								}
							} else {
								self.Errorf("While calculating possible sources for %v in %v: %v", email, params.GetString("GameId"), err)
							}
						}
					case "GetValidOrders":
						if loggedIn {
							params := subs.JSON{message.Method.Data}
							if options, err := game.GetValidOrders(self, params.GetString("GameId"), params.GetString("Province"), email); err == nil {
								if err := websocket.JSON.Send(ws, subs.Message{
									Type: common.RPCType,
									Method: &subs.Method{
										Name: message.Method.Name,
										Id:   message.Method.Id,
										Data: options,
									},
								}); err != nil {
									self.Errorf("%v", err)
								}
							} else {
								self.Errorf("While calculating valid orders for %v in %v in %v: %v", email, params.GetString("Province"), params.GetString("GameId"), err)
							}
						}
					case "SetOrder":
						if loggedIn {
							params := subs.JSON{message.Method.Data}
							result := game.SetOrder(self, params.GetString("GameId"), params.GetStringSlice("Order"), email)
							data := ""
							if result != nil {
								data = result.Error()
							}
							if err := websocket.JSON.Send(ws, subs.Message{
								Type: common.RPCType,
								Method: &subs.Method{
									Name: message.Method.Name,
									Id:   message.Method.Id,
									Data: data,
								},
							}); err != nil {
								self.Errorf("%v", err)
							}
						}
					}
				default:
					self.Errorf("Unrecognized message Type: %+v", message.Type)
				}
			}()
		} else if err == io.EOF {
			break
		} else {
			self.Errorf("%v", err)
		}
	}
}

func (self *Web) Openid(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	redirect, email, ok := openid.VerifyAuth(r)
	if ok {
		data.session.Values[SessionEmail] = email
		user.EnsureUser(self, email)
	} else {
		delete(data.session.Values, SessionEmail)
	}
	data.Close()
	w.Header().Set("Location", redirect.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect.String())
}

type token struct {
	Authorized bool
	Token      string
	Email      string
	Timeout    string
}

func (self *Web) hash(email string, timeout int64) string {
	hash := sha512.New()
	fmt.Fprintf(hash, "%s:%s:%s", email, self.secret, timeout)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func (self *Web) Token(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	if emailIf, found := data.session.Values[SessionEmail]; found {
		email := fmt.Sprint(emailIf)
		timeout := time.Now().Add(time.Minute).UnixNano()
		common.RenderJSON(w, token{
			Token:      self.hash(email, timeout),
			Authorized: true,
			Email:      email,
			Timeout:    fmt.Sprint(timeout),
		})
	} else {
		common.RenderJSON(w, token{
			Authorized: false,
		})
	}
}

func (self *Web) Logout(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	var redirect *url.URL
	r.ParseForm()
	if returnTo := r.Form.Get("return_to"); returnTo == "" {
		redirect = common.MustParseURL("http://" + r.Host + "/")
	} else {
		redirect = common.MustParseURL(returnTo)
	}
	delete(data.session.Values, SessionEmail)
	data.Close()
	w.Header().Set("Location", redirect.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, redirect.String())
}

func (self *Web) Login(w http.ResponseWriter, r *http.Request) {
	var redirect *url.URL
	r.ParseForm()
	if returnTo := r.Form.Get("return_to"); returnTo == "" {
		redirect = common.MustParseURL("http://" + r.Host + "/")
	} else {
		redirect = common.MustParseURL(returnTo)
	}
	url := openid.GetAuthURL(r, redirect)
	w.Header().Set("Location", url.String())
	w.WriteHeader(302)
	fmt.Fprintln(w, url.String())
}

func (self *Web) Index(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	common.SetContentType(w, "text/html; charset=UTF-8", false)
	self.renderText(w, r, self.htmlTemplates, "index.html", data)
}

func (self *Web) AppCache(w http.ResponseWriter, r *http.Request) {
	if self.appcache {
		data := self.GetRequestData(w, r)
		w.Header().Set("Content-Type", "AddType text/cache-manifest .appcache; charset=UTF-8")
		self.renderText(w, r, self.textTemplates, "diplicity.appcache", data)
	} else {
		w.WriteHeader(404)
	}
}

func (self *Web) AllJs(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	common.SetContentType(w, "application/javascript; charset=UTF-8", true)
	self.renderText(w, r, self.jsTemplates, "jquery-2.0.3.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "jquery.hammer.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "underscore-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "backbone-min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "bootstrap.min.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "log.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "util.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "panzoom.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "wsBackbone.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "baseView.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "dippyMap.js", data)
	fmt.Fprintln(w, ";")
	self.renderText(w, r, self.jsTemplates, "slider.js", data)
	fmt.Fprintln(w, ";")
	self.render_Templates(data)
	fmt.Fprintln(w, ";")
	for _, templ := range self.jsModelTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	for _, templ := range self.jsCollectionTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	for _, templ := range self.jsViewTemplates.Templates() {
		if err := templ.Execute(w, data); err != nil {
			panic(err)
		}
		fmt.Fprintln(w, ";")
	}
	self.renderText(w, r, self.jsTemplates, "app.js", data)
	fmt.Fprintln(w, ";")
}

func (self *Web) AllCss(w http.ResponseWriter, r *http.Request) {
	data := self.GetRequestData(w, r)
	w.Header().Set("Cache-Control", "public, max-age=864000")
	w.Header().Set("Content-Type", "text/css; charset=UTF-8")
	self.renderText(w, r, self.cssTemplates, "bootstrap.min.css", data)
	self.renderText(w, r, self.cssTemplates, "bootstrap-theme.min.css", data)
	self.renderText(w, r, self.cssTemplates, "slider.css", data)
	self.renderText(w, r, self.cssTemplates, "common.css", data)
}
