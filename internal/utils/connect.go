package utils

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/Xeway/amedee/internal/global"
)

// Connect to Hut Reservation service
// The client will have new cookies
func Connect(client *http.Client, username, password string) error {
	if username == "" || password == "" {
		return errors.New("email or password is empty")
	}

	csrfURL := global.BaseURL + "/api/v1/csrf"
	req, _ := http.NewRequest("GET", csrfURL, nil)
	req.Header.Set("User-Agent", "Amedee/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	xsrf := GetXSRFCookie(client)
	if xsrf == "" {
		return err
	}

	loginURL := global.BaseURL + "/api/v1/users/login"
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	loginReq, _ := http.NewRequest("POST", loginURL, bytes.NewBufferString(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("X-XSRF-TOKEN", xsrf)
	loginReq.Header.Set("Referer", global.BaseURL+"/login")
	loginReq.Header.Set("User-Agent", "Amedee/1.0")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

// GetXSRFCookie gets the XSRF token from the cookie jar
func GetXSRFCookie(client *http.Client) string {
	u, _ := url.Parse(global.BaseURL)
	for _, ck := range client.Jar.Cookies(u) {
		if ck.Name == "XSRF-TOKEN" {
			return ck.Value
		}
	}
	return ""
}
