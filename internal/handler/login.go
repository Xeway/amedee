package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	username := c.PostForm("email")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Please provide email & password"})
		return
	}

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 20 * 1e9}

	csrfURL := session.BaseURL + "/api/v1/csrf"
	req, _ := http.NewRequest("GET", csrfURL, nil)
	req.Header.Set("User-Agent", "Amedee/1.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("csrf request error:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream error (csrf)."})
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	xsrf := session.GetXSRFCookie(client)
	if xsrf == "" {
		log.Println("no xsrf cookie received")
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Missing XSRF token from upstream."})
		return
	}

	loginURL := session.BaseURL + "/api/v1/users/login"
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	loginReq, _ := http.NewRequest("POST", loginURL, bytes.NewBufferString(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("X-XSRF-TOKEN", xsrf)
	loginReq.Header.Set("Referer", session.BaseURL+"/login")
	loginReq.Header.Set("User-Agent", "Amedee/1.0")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Println("login call error:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream login error."})
		return
	}
	defer loginResp.Body.Close()
	bodyBytes, _ := io.ReadAll(loginResp.Body)

	if loginResp.StatusCode != http.StatusOK {
		log.Printf("login failed: status=%d body=%s\n", loginResp.StatusCode, string(bodyBytes))
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Invalid credentials."})
		return
	}

	if err := session.SaveClientCookiesToSession(c, client); err != nil {
		log.Println("failed to save cookies to session:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Internal server error."})
		return
	}

	c.HTML(http.StatusOK, "links.html", nil)
}
