package session

import (
	"bytes"
	"encoding/gob"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/Xeway/amedee/internal/global"
	"github.com/Xeway/amedee/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func init() {
	gob.Register([]*http.Cookie{})
}

// ClientFromSession builds an http.Client with cookies from session
func ClientFromSession(c *gin.Context) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(global.BaseURL)

	sess := sessions.Default(c)
	raw := sess.Get(global.SessionKey)
	if raw != nil {
		if raw == "anonymous" {
			// connect with an account
			client := &http.Client{Jar: jar, Timeout: 20 * 1e9}
			err := utils.Connect(client, os.Getenv("ANONYMOUS_USERNAME"), os.Getenv("ANONYMOUS_PASSWORD"))
			if err != nil {
				return nil, err
			}
		} else {
			if b, ok := raw.([]byte); ok {
				cs, err := BytesToCookies(b)
				if err != nil {
					return nil, err
				}
				jar.SetCookies(u, cs)
			}
		}
	}
	return &http.Client{Jar: jar, Timeout: 20 * time.Second}, nil
}

// SaveClientCookiesToSession saves cookies from client jar into session
func SaveClientCookiesToSession(c *gin.Context, client *http.Client, sessionDuration float64) error {
	u, _ := url.Parse(global.BaseURL)
	cookies := client.Jar.Cookies(u)
	b, err := CookiesToBytes(cookies)
	if err != nil {
		return err
	}
	sess := sessions.Default(c)
	sess.Set(global.SessionKey, b)
	sess.Options(sessions.Options{
		MaxAge: int(sessionDuration),
	})
	return sess.Save()
}

// SetSessionToAnonymous sets the session to anonymous
func SetSessionToAnonymous(c *gin.Context) error {
	sess := sessions.Default(c)
	sess.Set(global.SessionKey, "anonymous")
	return sess.Save()
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
