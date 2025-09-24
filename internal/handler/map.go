package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Map(c *gin.Context) {
	client, err := session.ClientFromSession(c)
	if err != nil {
		log.Println("clientFromSession err:", err)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	xsrf := session.GetXSRFCookie(client)
	if xsrf == "" {
		sess := sessions.Default(c)
		sess.Delete(session.SessionKey)
		_ = sess.Save()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	hutsURL := session.BaseURL + "/api/v1/manage/hutsList"
	req, _ := http.NewRequest("GET", hutsURL, nil)
	req.Header.Set("X-XSRF-TOKEN", xsrf)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("hutsList error:", err)
		c.String(http.StatusInternalServerError, "Upstream request error")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		sess := sessions.Default(c)
		sess.Delete(session.SessionKey)
		_ = sess.Save()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("read hutsList body error:", err)
		c.String(http.StatusInternalServerError, "Upstream read error")
		return
	}
	var huts interface{}
	if err := json.Unmarshal(body, &huts); err != nil {
		log.Println("unmarshal hutsList error:", err, "body=", string(body))
		c.String(http.StatusInternalServerError, "Invalid upstream data")
		return
	}

	c.HTML(http.StatusOK, "map.html", gin.H{"Huts": huts})
}
