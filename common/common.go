package common

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/user"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	regular = iota
	nilCache
)

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt64(s string) (result int64) {
	var err error
	if result, err = strconv.ParseInt(s, 10, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt(s string) (result int) {
	var err error
	if result, err = strconv.Atoi(s); err != nil {
		panic(err)
	}
	return
}

func AssertOkError(err error) {
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						panic(err)
					}
				}
			}
		} else if _, ok := err.(*datastore.ErrFieldMismatch); !ok {
			panic(err)
		}
	}
}

func SetContentType(w http.ResponseWriter, typ string) {
	w.Header().Set("Content-Type", typ)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
		panic(err)
	}
}

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score = MustParseFloat64(match[3])
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}

func isNil(v reflect.Value) bool {
	k := v.Kind()
	if k == reflect.Chan {
		return v.IsNil()
	}
	if k == reflect.Func {
		return v.IsNil()
	}
	if k == reflect.Interface {
		return v.IsNil()
	}
	if k == reflect.Map {
		return v.IsNil()
	}
	if k == reflect.Ptr {
		return v.IsNil()
	}
	if k == reflect.Slice {
		return v.IsNil()
	}
	return false
}

func keyify(k string) string {
	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	h := sha1.New()
	io.WriteString(h, k)
	sum := h.Sum(nil)
	if wrote, err := enc.Write(sum); err != nil {
		panic(err)
	} else if wrote != len(sum) {
		panic(fmt.Errorf("Tried to write %v bytes but wrote %v bytes", len(sum), wrote))
	}
	if err := enc.Close(); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}

func MemDel(c appengine.Context, keys ...string) {
	for index, key := range keys {
		keys[index] = keyify(key)
	}
	memcache.DeleteMulti(c, keys)
}

func Memoize2(c appengine.Context, super, key string, destP interface{}, f func() interface{}) (exists bool) {
	superH := keyify(super)
	var seed string
	item, err := memcache.Get(c, superH)
	if err != nil && err != memcache.ErrCacheMiss {
		panic(err)
	}
	if err == memcache.ErrCacheMiss {
		seed = fmt.Sprint(rand.Int63())
		item = &memcache.Item{
			Key:   superH,
			Value: []byte(seed),
		}
		c.Infof("Didn't find %v in memcache, reseeding with %v", super, seed)
		if err = memcache.Set(c, item); err != nil {
			panic(err)
		}
	} else {
		seed = string(item.Value)
	}
	return Memoize(c, fmt.Sprintf("%v@%v", key, seed), destP, f)
}

func reflectCopy(srcValue reflect.Value, source, destP interface{}) {
	if reflect.PtrTo(reflect.TypeOf(source)) == reflect.TypeOf(destP) {
		reflect.ValueOf(destP).Elem().Set(srcValue)
	} else {
		reflect.ValueOf(destP).Elem().Set(reflect.Indirect(srcValue))
	}
}

func Memoize(c appengine.Context, key string, destP interface{}, f func() interface{}) (exists bool) {
	return MemoizeMulti(c, []string{key}, []interface{}{destP}, []func() interface{}{f})[0]
}

func memGetMulti(c appengine.Context, keys []string, dests []interface{}) (items []*memcache.Item, errors []error) {
	items = make([]*memcache.Item, len(keys))
	errors = make([]error, len(keys))

	itemHash, err := memcache.GetMulti(c, keys)
	if err != nil {
		for index, _ := range errors {
			errors[index] = err
		}
		return
	}

	var item *memcache.Item
	var ok bool
	for index, keyHash := range keys {
		if item, ok = itemHash[keyHash]; ok {
			items[index] = item
			if err := memcache.Gob.Unmarshal(item.Value, dests[index]); err != nil {
				errors[index] = err
			}
		} else {
			errors[index] = memcache.ErrCacheMiss
		}
	}
	return
}

func MemoizeMulti(c appengine.Context, keys []string, destPs []interface{}, f []func() interface{}) (exists []bool) {
	exists = make([]bool, len(keys))
	keyHashes := make([]string, len(keys))
	for index, key := range keys {
		keyHashes[index] = keyify(key)
	}

	t := time.Now()
	items, errors := memGetMulti(c, keyHashes, destPs)
	if d := time.Now().Sub(t); d > time.Millisecond*10 {
		c.Debugf("SLOW memGetMulti(%v): %v", keys, d)
	}

	done := make(chan bool, len(items))

	for i, it := range items {
		index := i
		item := it
		err := errors[index]
		keyH := keyHashes[index]
		key := keys[index]
		destP := destPs[index]
		if err == memcache.ErrCacheMiss {
			c.Infof("Didn't find %v in memcache, regenerating", key)
			go func() {
				defer func() {
					done <- true
				}()
				result := f[index]()
				resultValue := reflect.ValueOf(result)
				if isNil(resultValue) {
					nilObj := reflect.Indirect(reflect.ValueOf(destP)).Interface()
					if err = memcache.Gob.Set(c, &memcache.Item{
						Key:    keyH,
						Flags:  nilCache,
						Object: nilObj,
					}); err != nil {
						c.Errorf("When trying to save %v(%v) => %#v: %#v", key, keyH, nilObj, err)
						panic(err)
					}
					exists[index] = false
				} else {
					if err = memcache.Gob.Set(c, &memcache.Item{
						Key:    keyH,
						Object: result,
					}); err != nil {
						c.Errorf("When trying to save %v(%v) => %#v: %#v", key, keyH, result, err)
						panic(err)
					} else {
						reflectCopy(resultValue, result, destP)
						exists[index] = true
					}
				}
			}()
		} else if err != nil {
			c.Errorf("When trying to get %v(%v): %#v", key, keyH, err)
			panic(err)
		} else {
			if item.Flags&nilCache == nilCache {
				exists[index] = false
			} else {
				exists[index] = true
			}
			done <- true
		}
	}
	for i := 0; i < len(items); i++ {
		<-done
	}
	return
}

func HostURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v", scheme, r.Host)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	SetContentType(w, "application/json; charset=UTF-8")
	MustEncodeJSON(w, data.User)
}

type RequestData struct {
	Response     http.ResponseWriter
	Request      *http.Request
	Context      appengine.Context
	User         *user.User
	translations map[string]string
}

func GetRequestData(w http.ResponseWriter, r *http.Request) (result RequestData) {
	result = RequestData{
		Response:     w,
		Request:      r,
		Context:      appengine.NewContext(r),
		translations: getTranslations(r),
	}
	result.User = user.Current(result.Context)
	return
}

func (self RequestData) LogoutURL() string {
	result, err := user.LogoutURL(self.Context, HostURL(self.Request))
	if err != nil {
		panic(err)
	}
	return result
}

func (self RequestData) LoginURL() string {
	result, err := user.LoginURL(self.Context, HostURL(self.Request))
	if err != nil {
		panic(err)
	}
	return result
}

func (self RequestData) Authenticated() bool {
	if self.User == nil {
		self.Response.WriteHeader(401)
		fmt.Fprintln(self.Response, "Unauthorized")
		return false
	}
	return true
}

func (self RequestData) I(phrase string, args ...string) string {
	pattern, ok := self.translations[phrase]
	if !ok {
		panic(fmt.Errorf("Found no translation for %v", phrase))
	}
	if len(args) > 0 {
		return fmt.Sprintf(pattern, args)
	}
	return pattern
}

var debugVersion time.Time

func (self RequestData) Version() string {
	if appengine.IsDevAppServer() {
		if debugVersion.Before(time.Now().Add(-time.Second)) {
			debugVersion = time.Now()
		}
		return fmt.Sprintf("%v.%v", appengine.VersionID(self.Context), debugVersion.UnixNano())
	}
	return appengine.VersionID(self.Context)
}

func (self RequestData) Inline(p string) string {
	in, err := os.Open(p)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	return string(b)
}
