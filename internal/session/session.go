package session

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/Xeway/amedee/internal/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
func SaveClientCookiesToSession(c *gin.Context, client *http.Client) error {
	u, _ := url.Parse(BaseURL)
	cookies := client.Jar.Cookies(u)
	b, err := CookiesToBytes(cookies)
	if err != nil {
		return err
	}
	sess := sessions.Default(c)
	sess.Set(SessionKey, b)
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
	s := make([]model.SerializableCookie, 0, len(cookies))
	for _, c := range cookies {
		s = append(s, model.SerializableCookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.Expires,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		})
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BytesToCookies deserializes cookies from session
func BytesToCookies(b []byte) ([]*http.Cookie, error) {
	var s []model.SerializableCookie
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&s); err != nil {
		return nil, err
	}
	out := make([]*http.Cookie, 0, len(s))
	for _, sc := range s {
		out = append(out, &http.Cookie{
			Name:     sc.Name,
			Value:    sc.Value,
			Path:     sc.Path,
			Domain:   sc.Domain,
			Expires:  sc.Expires,
			Secure:   sc.Secure,
			HttpOnly: sc.HttpOnly,
		})
	}
	return out, nil
}
