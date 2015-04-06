package user

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zond/diplicity/srv"
	"github.com/zond/goauth2"
	"github.com/zond/unbolted"
	"github.com/zond/wsubs/gosubs"
)

func AdminSetRank1(c *srv.HTTPContext) (err error) {
	return c.DB().Update(func(tx *unbolted.TX) (err error) {
		users := Users{}
		if err = tx.Query().All(&users); err != nil {
			return
		}
		for _, user := range users {
			user.Ranking = 1
			if err = tx.Set(&user); err != nil {
				return
			}
			fmt.Fprintf(c.Resp(), "Set rank of %#v to 1\n", user.Email)
		}
		return
	})
}

func DevLogin(c *srv.HTTPContext) (err error) {
	c.SetContentType("text/html; charset=UTF-8")
	f, err := os.Open(filepath.Join("static", "devlogin.html"))
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = io.Copy(c.Resp(), f); err != nil {
		return
	}
	return
}

func DevBecome(c *srv.HTTPContext) (err error) {
	email := c.Req().FormValue("become")
	if err = c.DB().Update(func(tx *unbolted.TX) (err error) {
		user := &User{Id: unbolted.Id(email)}
		err = tx.Get(user)
		if err == unbolted.ErrNotFound {
			user.Nickname = email
			user.Email = email
			user.Ranking = 1
			user.LastLoginAt = time.Now()
			user.DiplicityHost = c.Req().Host
			return tx.Set(user)
		}
		return err
	}); err != nil {
		return
	}
	c.Session().Values[srv.SessionEmail] = c.Req().FormValue("become")
	c.Close()
	http.Redirect(c.Resp(), c.Req(), "/", 302)
	return
}

func AdminCreateUser(c *srv.HTTPContext) (err error) {
	user := &User{}
	if err = json.NewDecoder(c.Req().Body).Decode(user); err != nil {
		return
	}
	return c.DB().Update(func(tx *unbolted.TX) error { return tx.Set(user) })
}

var nonces = map[string]struct{}{}
var nonceLock = sync.Mutex{}

func OAuth2Callback(clientId, clientSecret string) func(c *srv.HTTPContext) (err error) {
	return func(c *srv.HTTPContext) (err error) {
		state := c.Req().FormValue("state")
		parts := strings.SplitN(state, ".", 2)
		if len(parts) != 2 {
			err = fmt.Errorf("state %#v is invalid", state)
			return
		}
		nonceLock.Lock()
		defer nonceLock.Unlock()
		if _, found := nonces[parts[0]]; !found {
			err = fmt.Errorf("nonce not found")
			return
		}
		delete(nonces, parts[0])

		returnTo, err := url.Parse(parts[1])
		if err != nil {
			return
		}
		redirectUrl, err := url.Parse(fmt.Sprintf("%v://%v/oauth2callback", returnTo.Scheme, c.Req().Host))
		if err != nil {
			return
		}
		email, ok, err := goauth2.VerifyEmail(clientId, clientSecret, c.Req().FormValue("code"), redirectUrl)
		if err != nil {
			return
		}

		if ok {
			email = strings.ToLower(email)
			c.Session().Values[srv.SessionEmail] = email
			if err = c.DB().Update(func(tx *unbolted.TX) (err error) {
				u := &User{Id: unbolted.Id(email)}
				err = tx.Get(u)
				if err == unbolted.ErrNotFound {
					err = nil
					u.Email = email
					u.Ranking = 1
				}
				if err == nil {
					u.DiplicityHost = c.Req().Host
					u.LastLoginAt = time.Now()
					err = tx.Set(u)
				}
				return
			}); err != nil {
				return
			}
		} else {
			delete(c.Session().Values, srv.SessionEmail)
		}
		c.Close()
		c.Resp().Header().Set("Location", returnTo.String())
		c.Resp().WriteHeader(302)
		fmt.Fprintln(c.Resp(), returnTo.String())
		return
	}
}

func Token(c *srv.HTTPContext) (err error) {
	if emailIf, found := c.Session().Values[srv.SessionEmail]; found {
		token := &gosubs.Token{
			Principal: fmt.Sprint(emailIf),
			Timeout:   time.Now().Add(time.Second * 10),
		}
		if err = token.Encode(c.Secret()); err != nil {
			return
		}
		err = c.RenderJSON(token)
	} else {
		err = c.RenderJSON(gosubs.Token{})
	}
	return
}

func Logout(c *srv.HTTPContext) (err error) {
	delete(c.Session().Values, srv.SessionEmail)
	c.Close()
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	c.Resp().Header().Set("Location", redirect)
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), redirect)
	return
}

func Login(clientId string) func(c *srv.HTTPContext) (err error) {
	return func(c *srv.HTTPContext) (err error) {
		returnTo, err := url.Parse(c.Req().FormValue("return_to"))
		if err != nil {
			return
		}

		if c.Env() == srv.Development {
			http.Redirect(c.Resp(), c.Req(), "/admin/login", 302)
			return
		}

		redirectUrl, err := url.Parse(fmt.Sprintf("%v://%v/oauth2callback", returnTo.Scheme, c.Req().Host))
		if err != nil {
			return
		}
		nonce := fmt.Sprintf("%x%x", rand.Int63(), rand.Int63())
		nonceLock.Lock()
		defer nonceLock.Unlock()
		nonces[nonce] = struct{}{}
		nonce += "." + returnTo.String()
		url, err := goauth2.GetAuthURL(clientId, nonce, redirectUrl)
		if err != nil {
			return
		}
		http.Redirect(c.Resp(), c.Req(), url.String(), http.StatusTemporaryRedirect)
		return
	}
}
