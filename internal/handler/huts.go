package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Xeway/amedee/internal/session"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AvailabilityInfo represents the structure of a single day's availability from the API
type AvailabilityInfo struct {
	FreeBeds int       `json:"freeBeds"`
	Date     time.Time `json:"date"`
}

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

	// --- New: Parameter Handling for Availability Check ---
	var checkAvailability bool
	var startDate, endDate time.Time
	var numPeople int

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	numPeopleStr := c.Query("numPeople")

	if startDateStr != "" && endDateStr != "" && numPeopleStr != "" {
		const layout = "2006-01-02"
		startDate, err = time.Parse(layout, startDateStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid startDate format. Use YYYY-MM-DD.")
			return
		}
		endDate, err = time.Parse(layout, endDateStr)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid endDate format. Use YYYY-MM-DD.")
			return
		}
		if endDate.Before(startDate) {
			c.String(http.StatusBadRequest, "endDate cannot be before startDate.")
			return
		}

		numPeople, err = strconv.Atoi(numPeopleStr)
		if err != nil || numPeople < 1 {
			c.String(http.StatusBadRequest, "Invalid numPeople. Must be a positive integer.")
			return
		}

		checkAvailability = true
	}
	// --- End of New Parameter Handling ---

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

	var hutsList []map[string]interface{}
	if err := json.Unmarshal(body, &hutsList); err != nil {
		log.Println("unmarshal hutsList error:", err, "body=", string(body))
		c.String(http.StatusInternalServerError, "Invalid upstream data")
		return
	}

	var filteredHuts []map[string]interface{}
	for _, hut := range hutsList {
		if hut["hutCountry"] == "CH" {
			filteredHuts = append(filteredHuts, hut)
		}
	}
	hutsList = filteredHuts

	results := make([]map[string]interface{}, len(hutsList))
	for i := range hutsList {
		m := make(map[string]interface{}, len(hutsList[i]))
		for k, v := range hutsList[i] {
			m[k] = v
		}
		results[i] = m
	}

	const maxConcurrency = 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, hut := range hutsList {
		idVal, ok := hut["hutId"]
		if !ok {
			continue
		}

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

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout for multiple requests
			defer cancel()

			// --- Request 1: Get Hut Info ---
			hutInfoURL := session.BaseURL + "/api/v1/reservation/hutInfo/" + id
			reqInfo, _ := http.NewRequestWithContext(ctx, "GET", hutInfoURL, nil)
			reqInfo.Header.Set("X-XSRF-TOKEN", xsrf)
			reqInfo.Header.Set("Accept", "application/json")

			var hutInfo map[string]interface{}
			respInfo, err := client.Do(reqInfo)
			if err != nil {
				log.Println("hutInfo error for hutId", id, ":", err)
				return // Continue to next hut
			}
			defer respInfo.Body.Close()

			if respInfo.StatusCode == http.StatusUnauthorized {
				mu.Lock()
				sess := sessions.Default(c)
				sess.Delete(session.SessionKey)
				_ = sess.Save()
				mu.Unlock()
				return
			}

			bodyInfo, err := io.ReadAll(respInfo.Body)
			if err != nil {
				log.Println("read hutInfo body err for hutId", id, ":", err)
			} else {
				if err := json.Unmarshal(bodyInfo, &hutInfo); err != nil {
					log.Println("unmarshal hutInfo error for hutId", id, ":", err, "body=", string(bodyInfo))
				}
			}

			// --- Request 2: Get Hut Availability ---
			isAvailable := false // Default to false if check is enabled
			if checkAvailability {
				availURL := fmt.Sprintf("%s/api/v1/reservation/getHutAvailability?hutId=%s&step=WIZARD", session.BaseURL, id)
				reqAvail, _ := http.NewRequestWithContext(ctx, "GET", availURL, nil)
				reqAvail.Header.Set("X-XSRF-TOKEN", xsrf)
				reqAvail.Header.Set("Accept", "application/json")

				respAvail, err := client.Do(reqAvail)
				if err != nil {
					log.Printf("hutAvailability error for hutId %s: %v", id, err)
				} else {
					defer respAvail.Body.Close()
					if respAvail.StatusCode == http.StatusOK {
						bodyAvail, err := io.ReadAll(respAvail.Body)
						if err != nil {
							log.Printf("read hutAvailability body err for hutId %s: %v", id, err)
						} else {
							var availabilityData []AvailabilityInfo
							if err := json.Unmarshal(bodyAvail, &availabilityData); err != nil {
								log.Printf("unmarshal hutAvailability err for hutId %s: %v, body=%s", id, err, string(bodyAvail))
							} else {
								// Process the data to check availability for the date range
								isAvailable = isHutAvailableForRange(availabilityData, startDate, endDate, numPeople)
							}
						}
					}
				}
			}

			// --- Merge all results under a lock ---
			mu.Lock()
			defer mu.Unlock()

			// Merge info from the first request
			for k, v := range hutInfo {
				results[idx][k] = v
			}

			// Add availability status from the second request
			if checkAvailability {
				results[idx]["isAvailable"] = isAvailable
			}
		}(i, hutIDStr)
	}

	wg.Wait()

	c.JSON(http.StatusOK, results)
}

// isHutAvailableForRange checks if a hut has enough free beds for the entire date range.
func isHutAvailableForRange(availabilityData []AvailabilityInfo, start, end time.Time, numPeople int) bool {
	// Create a map for quick lookup of free beds by date.
	bedsByDate := make(map[string]int)
	for _, day := range availabilityData {
		// Key by "YYYY-MM-DD"
		dateKey := day.Date.Format("2006-01-02")
		bedsByDate[dateKey] = day.FreeBeds
	}

	// Iterate through every day in the requested range.
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		freeBeds, ok := bedsByDate[dateKey]

		// If a day is missing from the data OR has too few beds, the hut is not available.
		if !ok || freeBeds < numPeople {
			return false
		}
	}

	// If all days in the range have enough beds, the hut is available.
	return true
}
