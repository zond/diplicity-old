package common

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

const (
	discovery = "https://www.google.com/accounts/o8/id"
)

var endpoint *url.URL

type xrdDoc struct {
	XRD string `xml:"XRD>Service>URI"`
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
	var err error
	if result, err = url.Parse(endp.String() + "?" + query.Encode()); err != nil {
		panic(err)
	}
	return
}
