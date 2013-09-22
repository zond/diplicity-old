package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zond/diplicity/web"
	"net/http"
	"net/url"
)

func main() {
	host := flag.String("host", "localhost", "The host to connect to")
	port := flag.Int("port", 8080, "The port to connect to")
	email := flag.String("email", "", "The email to fake authenticating as")
	secret := flag.String("secret", web.DefaultSecret, "The cookie store secret of the server")

	flag.Parse()

	location, err := url.Parse(fmt.Sprintf("ws://%v:%v/ws", *host, *port))
	if err != nil {
		panic(err)
	}
	origin, err := url.Parse("http://localhost")
	if err != nil {
		panic(err)
	}

	encoded, err := securecookie.EncodeMulti(web.SessionName, map[interface{}]interface{}{
		web.SessionEmail: *email,
	}, securecookie.CodecsFromPairs([]byte(*secret))...)
	if err != nil {
		panic(err)
	}

	cookie := sessions.NewCookie(web.SessionName, encoded, &sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 30,
	})

	header := http.Header{
		"Cookie": []string{cookie.String()},
	}

	config := &websocket.Config{
		Location: location,
		Origin:   origin,
		Version:  13,
		Header:   header,
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		return
	}
}
