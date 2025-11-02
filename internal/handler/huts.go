package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Huts(c *gin.Context) {
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

	// Expect an array of hut objects: [{"hutName":"...","hutId":603,...}, ...]
	var hutsList []map[string]interface{}
	if err := json.Unmarshal(body, &hutsList); err != nil {
		log.Println("unmarshal hutsList error:", err, "body=", string(body))
		c.String(http.StatusInternalServerError, "Invalid upstream data")
		return
	}

	// Prepare results by copying original items
	results := make([]map[string]interface{}, len(hutsList))
	for i := range hutsList {
		// shallow copy map to avoid races when merging
		m := make(map[string]interface{}, len(hutsList[i]))
		for k, v := range hutsList[i] {
			m[k] = v
		}
		results[i] = m
	}

	// Concurrency control
	const maxConcurrency = 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, hut := range hutsList {
		// extract hutId
		idVal, ok := hut["hutId"]
		if !ok {
			continue
		}

		// normalize hutId to string
		var hutIDStr string
		switch v := idVal.(type) {
		case float64:
			hutIDStr = fmt.Sprintf("%.0f", v)
		case string:
			hutIDStr = v
		case json.Number:
			hutIDStr = v.String()
		default:
			hutIDStr = fmt.Sprintf("%v", v)
		}

		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// use a short timeout per hut info request
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()

			hutInfoURL := session.BaseURL + "/api/v1/reservation/hutInfo/" + id
			req, _ := http.NewRequestWithContext(ctx, "GET", hutInfoURL, nil)
			req.Header.Set("X-XSRF-TOKEN", xsrf)
			req.Header.Set("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Println("hutInfo error for hutId", id, ":", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				// If upstream revoked auth, clear session and stop
				mu.Lock()
				sess := sessions.Default(c)
				sess.Delete(session.SessionKey)
				_ = sess.Save()
				mu.Unlock()
				return
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("read hutInfo body err for hutId", id, ":", err)
				return
			}

			var info map[string]interface{}
			if err := json.Unmarshal(b, &info); err != nil {
				log.Println("unmarshal hutInfo error for hutId", id, ":", err, "body=", string(b))
				return
			}

			// Merge info into results[idx]
			mu.Lock()
			for k, v := range info {
				results[idx][k] = v
			}
			mu.Unlock()
		}(i, hutIDStr)
	}

	wg.Wait()

	c.JSON(http.StatusOK, results)
}
