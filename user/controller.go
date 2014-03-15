package user

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/zond/diplicity/common"
	"github.com/zond/gopenid"
	"github.com/zond/wsubs/gosubs"
)

func Openid(c *common.HTTPContext) (err error) {
	redirect, email, ok, err := gopenid.VerifyAuth(c.Req())
	if err != nil {
		return
	}
	if ok {
		c.Session().Values[common.SessionEmail] = strings.ToLower(email)
		EnsureUser(c.DB(), email)
	} else {
		delete(c.Session().Values, common.SessionEmail)
	}
	c.Close()
	c.Resp().Header().Set("Location", redirect.String())
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), redirect.String())
	return
}

func Token(c *common.HTTPContext) (err error) {
	if emailIf, found := c.Session().Values[common.SessionEmail]; found {
		token := &gosubs.Token{
			Principal: fmt.Sprint(emailIf),
			Timeout:   time.Now().Add(time.Second * 10),
		}
		if err = token.Encode(); err != nil {
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

func Login(c *common.HTTPContext) (err error) {
	redirect := c.Req().FormValue("return_to")
	if redirect == "" {
		redirect = fmt.Sprintf("http://%v/", c.Req().Host)
	}
	redirectUrl, err := url.Parse(redirect)
	if err != nil {
		return
	}
	url, err := gopenid.GetAuthURL(c.Req(), redirectUrl)
	if err != nil {
		return
	}
	c.Resp().Header().Set("Location", url.String())
	c.Resp().WriteHeader(302)
	fmt.Fprintln(c.Resp(), url.String())
	return
}
