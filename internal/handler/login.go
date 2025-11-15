package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"

	"github.com/Xeway/amedee/internal/global"
	"github.com/Xeway/amedee/internal/session"
	"github.com/Xeway/amedee/internal/utils"
	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	wantsAnonymous := c.PostForm("anonymous") == "1"
	if wantsAnonymous {
		if err := session.SetSessionToAnonymous(c); err != nil {
			log.Println("failed to save cookie to session:", err)
			c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Internal server error."})
			return
		}
		c.HTML(http.StatusOK, "links.html", nil)
		return
	}

	username := c.PostForm("email")
	password := c.PostForm("password")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 20 * 1e9}

	err := utils.Connect(client, username, password)
	if err != nil {
		log.Println("login error:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": err.Error()})
		return
	}

	// Get current user to retrieve session duration
	var userInfo map[string]interface{}

	currentUserUrl := global.BaseURL + "/api/v1/manage/currentUser"
	req, _ := http.NewRequest("GET", currentUserUrl, nil)
	req.Header.Set("User-Agent", "Amedee/1.0")

	respUser, err := client.Do(req)
	if err != nil {
		log.Println("current user request error:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream error (current user)."})
		return
	}
	io.Copy(io.Discard, respUser.Body)
	respUser.Body.Close()
	if respUser.StatusCode != http.StatusOK {
		log.Printf("current user request failed with status code %d", respUser.StatusCode)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream error (current user)."})
		return
	}
	bodyUser, err := io.ReadAll(respUser.Body)
	if err != nil {
		log.Println("current user read error:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream error (current user)."})
		return
	} else {
		if err := json.Unmarshal(bodyUser, &userInfo); err != nil {
			log.Println("current user unmarshal error:", err)
			c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Upstream error (current user)."})
			return
		}
	}
	sessionDuration := userInfo["sessionTimeout"].(float64)

	if err := session.SaveClientCookiesToSession(c, client, sessionDuration); err != nil {
		log.Println("failed to save cookies to session:", err)
		c.HTML(http.StatusOK, "login_modal.html", gin.H{"Error": "Internal server error."})
		return
	}

	c.HTML(http.StatusOK, "links.html", nil)
}
