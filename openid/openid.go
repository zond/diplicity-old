package openid

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/zond/diplicity/common"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	discovery    = "https://www.google.com/accounts/o8/id"
	maxOldNonces = 100000
)

type oldNonce struct {
	at    time.Time
	nonce string
	next  *oldNonce
	prev  *oldNonce
}

type oldNonces struct {
	nonceMap       map[string]*oldNonce
	nonceListStart *oldNonce
	nonceListEnd   *oldNonce
	max            int
}

func newOldNonces() *oldNonces {
	return &oldNonces{
		nonceMap: make(map[string]*oldNonce),
		max:      maxOldNonces,
	}
}

func (self *oldNonces) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", self.nonceMap)
	for n := self.nonceListStart; n != nil; n = n.next {
		fmt.Fprintf(buf, "%v@%v =>\n", n.nonce, n.at)
	}
	return string(buf.Bytes())
}

func (self *oldNonces) size() int {
	return len(self.nonceMap)
}

func (self *oldNonces) add(s string) bool {
	if _, found := self.nonceMap[s]; found {
		return false
	}
	n := &oldNonce{
		at:    time.Now(),
		nonce: s,
		next:  self.nonceListStart,
	}
	self.nonceListStart = n
	if n.next != nil {
		n.next.prev = n
	}
	if self.nonceListEnd == nil {
		self.nonceListEnd = self.nonceListStart
	}
	self.nonceMap[s] = self.nonceListStart
	for len(self.nonceMap) > self.max {
		last := self.nonceListEnd
		last.prev.next = nil
		self.nonceListEnd = last.prev
		delete(self.nonceMap, last.nonce)
	}
	return true
}

var nonces = newOldNonces()
var endpoint *url.URL

type xrdDoc struct {
	XRD string `xml:"XRD>Service>URI"`
}

func join(u *url.URL, q url.Values) (result *url.URL) {
	buf := bytes.NewBufferString(u.Scheme)
	fmt.Fprint(buf, "://")
	if u.User != nil {
		fmt.Fprintf(buf, "%v@", u.User.String())
	}
	fmt.Fprint(buf, u.Host)
	if u.Path != "" {
		fmt.Fprint(buf, u.Path)
	}
	if u.Fragment != "" {
		fmt.Fprintf(buf, "#%v", u.Fragment)
	}
	if u.RawQuery == "" {
		fmt.Fprintf(buf, "?%v", q.Encode())
	} else {
		fmt.Fprintf(buf, "?%v&%v", u.RawQuery, q.Encode())
	}
	result = common.MustParseURL(string(buf.Bytes()))
	return
}

func getEndpoint() *url.URL {
	var err error
	if endpoint == nil {
		var req *http.Request
		if req, err = http.NewRequest("GET", discovery, nil); err != nil {
			panic(err)
		}
		var resp *http.Response
		if resp, err = new(http.Client).Do(req); err != nil {
			panic(err)
		}
		dec := xml.NewDecoder(resp.Body)
		var x xrdDoc
		if err = dec.Decode(&x); err != nil {
			panic(err)
		}
		endpoint = common.MustParseURL(x.XRD)
	}
	return endpoint
}

func VerifyAuth(r *http.Request) (returnTo *url.URL, result string, ok bool) {
	endp := getEndpoint()
	query := endp.Query()
	r.ParseForm()
	var nonce string
	for key, values := range r.Form {
		for _, value := range values {
			if key == "openid.ext1.value.email" {
				result = value
			}
			if key == "openid.secondary_return_to" {
				returnTo = common.MustParseURL(value)
			}
			if key == "openid.response_nonce" {
				nonce = value
			}
			query.Add(key, value)
		}
	}
	query.Set("openid.mode", "check_authentication")
	response, err := new(http.Client).Get(join(endp, query).String())
	if err != nil {
		panic(err)
	}
	bod, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(bod), "\n") {
		kv := strings.SplitN(line, ":", 2)
		switch kv[0] {
		case "is_valid":
			if kv[1] == "true" {
				if nonces.add(nonce) {
					ok = true
				}
			}
		case "ns":
			if kv[1] != "http://specs.openid.net/auth/2.0" {
				panic(fmt.Errorf("Unknown namespace: %v", kv[1]))
			}
		}
	}
	return
}

func GetAuthURL(r *http.Request, returnTo *url.URL) (result *url.URL) {
	endp := getEndpoint()
	query := endp.Query()
	query.Add("openid.mode", "checkid_setup")
	query.Add("openid.ns", "http://specs.openid.net/auth/2.0")
	query.Add("openid.return_to", "http://"+r.Host+"/openid?openid.secondary_return_to="+returnTo.String())
	query.Add("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	query.Add("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	query.Add("openid.ns.ax", "http://openid.net/srv/ax/1.0")
	query.Add("openid.ax.mode", "fetch_request")
	query.Add("openid.ax.required", "email")
	query.Add("openid.ax.type.email", "http://axschema.org/contact/email")
	return join(endp, query)
}
