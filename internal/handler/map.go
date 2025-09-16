package handler

import (
	"encoding/json"
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

	b := make([]byte, 0)
	_, _ = resp.Body.Read(b)
	var huts interface{}
	_ = json.Unmarshal(b, &huts)

	c.HTML(http.StatusOK, "map.html", gin.H{"Huts": huts})
}
