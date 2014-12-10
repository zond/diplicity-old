package user

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/goauth2"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/wsubs/gosubs"
)

func AdminSetRank1(c *common.HTTPContext) (err error) {
	users := Users{}
	if err = c.DB().Query().All(&users); err != nil {
		return
	}
	for _, user := range users {
		user.Ranking = 1
		if err = c.DB().Set(&user); err != nil {
			return
		}
		fmt.Fprintf(c.Resp(), "Set rank of %#v to 1\n", user.Email)
	}
	return

}

func AdminBecome(c *common.HTTPContext) (err error) {
	c.Session().Values[common.SessionEmail] = c.Req().FormValue("become")
	c.Close()
	c.Resp().Header().Set("Location", "/")
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), "/")
	return
}

func AdminCreateUser(c *common.HTTPContext) (err error) {
	user := &User{}
	if err = json.NewDecoder(c.Req().Body).Decode(user); err != nil {
		return
	}
	err = c.DB().Set(user)
	return
}

var nonces = map[string]struct{}{}
var nonceLock = sync.Mutex{}

func OAuth2Callback(clientId, clientSecret string) func(c *common.HTTPContext) (err error) {
	return func(c *common.HTTPContext) (err error) {
		state := c.Req().FormValue("state")
		nonceLock.Lock()
		defer nonceLock.Unlock()
		if _, found := nonces[state]; !found {
			err = fmt.Errorf("state not found")
			return
		}
		delete(nonces, state)

		scheme := "http"
		if c.Req().TLS != nil {
			scheme = "https"
		}
		redirectUrl, err := url.Parse(fmt.Sprintf("%v://%v/oauth2callback", scheme, c.Req().Host))
		if err != nil {
			return
		}
		email, ok, err := goauth2.VerifyEmail(clientId, clientSecret, c.Req().FormValue("code"), redirectUrl)
		if err != nil {
			return
		}

		if ok {
			email = strings.ToLower(email)
			c.Session().Values[common.SessionEmail] = email
			u := &User{Id: kol.Id(email)}
			err = c.DB().Get(u)
			if err == kol.NotFound {
				err = nil
				u.Email = email
				u.Ranking = 1
			}
			if err == nil {
				u.Language = common.GetLanguage(c.Req())
				u.DiplicityHost = c.Req().Host
				u.LastLoginAt = time.Now()
				err = c.DB().Set(u)
			}
		} else {
			delete(c.Session().Values, common.SessionEmail)
		}
		c.Close()
		c.Resp().Header().Set("Location", "/")
		c.Resp().WriteHeader(302)
		fmt.Fprintln(c.Resp(), "/")
		return
	}
}

func Token(c *common.HTTPContext) (err error) {
	if emailIf, found := c.Session().Values[common.SessionEmail]; found {
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

func Logout(c *common.HTTPContext) (err error) {
	delete(c.Session().Values, common.SessionEmail)
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

func Login(clientId string) func(c *common.HTTPContext) (err error) {
	return func(c *common.HTTPContext) (err error) {
		scheme := "http"
		if c.Req().TLS != nil {
			scheme = "https"
		}
		redirectUrl, err := url.Parse(fmt.Sprintf("%v://%v/oauth2callback", scheme, c.Req().Host))
		if err != nil {
			return
		}
		nonce := fmt.Sprintf("%x%x", rand.Int63(), rand.Int63())
		nonceLock.Lock()
		defer nonceLock.Unlock()
		nonces[nonce] = struct{}{}
		url, err := goauth2.GetAuthURL(clientId, nonce, redirectUrl)
		if err != nil {
			return
		}
		c.Resp().Header().Set("Location", url.String())
		c.Resp().WriteHeader(302)
		fmt.Fprintln(c.Resp(), url.String())
		return
	}
}
