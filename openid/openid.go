package openid

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	discovery = "https://www.google.com/accounts/o8/id"
)

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
	var err error
	if result, err = url.Parse(string(buf.Bytes())); err != nil {
		panic(err)
	}
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
		if endpoint, err = url.Parse(x.XRD); err != nil {
			panic(err)
		}
	}
	return endpoint
}

func VerifyAuth(r *http.Request) (result string, ok bool) {
	endp := getEndpoint()
	query := endp.Query()
	r.ParseForm()
	for key, values := range r.Form {
		for _, value := range values {
			if key == "openid.ext1.value.email" {
				result = value
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
				ok = true
			}
		case "ns":
			if kv[1] != "http://specs.openid.net/auth/2.0" {
				panic(fmt.Errorf("Unknown namespace: %v", kv[1]))
			}
		}
	}
	return
}

func GetAuthURL(r *http.Request) (result *url.URL) {
	endp := getEndpoint()
	query := endp.Query()
	query.Add("openid.mode", "checkid_setup")
	query.Add("openid.ns", "http://specs.openid.net/auth/2.0")
	query.Add("openid.return_to", "http://"+r.Host+"/openid")
	query.Add("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	query.Add("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	query.Add("openid.ns.ax", "http://openid.net/srv/ax/1.0")
	query.Add("openid.ax.mode", "fetch_request")
	query.Add("openid.ax.required", "email")
	query.Add("openid.ax.type.email", "http://axschema.org/contact/email")
	return join(endp, query)
}
