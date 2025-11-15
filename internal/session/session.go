package session

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register([]*http.Cookie{})
}

const (
	BaseURL    = "https://www.hut-reservation.org"
	SessionKey = "thirdparty_cookies"
)

// ClientFromSession builds an http.Client with cookies from session
func ClientFromSession(c *gin.Context) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(BaseURL)

	sess := sessions.Default(c)
	raw := sess.Get(SessionKey)
	if raw != nil {
		if b, ok := raw.([]byte); ok {
			cs, err := BytesToCookies(b)
			if err != nil {
				return nil, err
			}
			jar.SetCookies(u, cs)
		}
	}
	return &http.Client{Jar: jar, Timeout: 20 * time.Second}, nil
}

// SaveClientCookiesToSession saves cookies from client jar into session
func SaveClientCookiesToSession(c *gin.Context, client *http.Client, sessionDuration float64) error {
	u, _ := url.Parse(BaseURL)
	cookies := client.Jar.Cookies(u)
	b, err := CookiesToBytes(cookies)
	if err != nil {
		return err
	}
	sess := sessions.Default(c)
	sess.Set(SessionKey, b)
	sess.Options(sessions.Options{
		MaxAge: int(sessionDuration),
	})
	return sess.Save()
}

// GetXSRFCookie gets the XSRF token from the cookie jar
func GetXSRFCookie(client *http.Client) string {
	u, _ := url.Parse(BaseURL)
	for _, ck := range client.Jar.Cookies(u) {
		if ck.Name == "XSRF-TOKEN" {
			return ck.Value
		}
	}
	return ""
}

// CookiesToBytes serializes cookies for session storage
func CookiesToBytes(cookies []*http.Cookie) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(cookies); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BytesToCookies deserializes cookies from session
func BytesToCookies(b []byte) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&cookies); err != nil {
		return nil, err
	}
	return cookies, nil
}
