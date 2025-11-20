package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c Controller) GetNext(w http.ResponseWriter, r *http.Request) {
	streetName := parseStreetName(r)
	houseNumber := parseHouseNumber(r)

	redirectUrl, err := c.Client.GetRedirectUrl(InitialContextPath)

	if err != nil {
		// TODO handle 404
		c.createInternalServerError(w, err)
		return
	}

	var csvResponse []byte

	csvResponse, err = c.Client.GetCSV(redirectUrl, url.QueryEscape(streetName), houseNumber)

	if err != nil || csvResponse == nil {
		c.createInternalServerError(w, err)
		return
	}

	csvGarbageLines := strings.Split(string(csvResponse), "\n")[1:]
	nearestNextDate := findNearestNextDate(csvGarbageLines)
	//fmt.Println("Nearest date is ", nearestNextDate)
	garbageTypes := getGarbageTypes(csvGarbageLines, nearestNextDate)

	response := nextResponse{
		Date:  nearestNextDate.Format("2006-01-02"),
		Types: garbageTypes,
	}

	// JSON response with proper content type and UTF-8 charset
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func findNearestNextDate(csvEntries []string) time.Time {
	now := time.Now()
	//fmt.Println("Current date and time is: ", now.String())

	nearestDate, _ := time.Parse("02.01.2006", "31.12.2100")

	for _, csvLine := range csvEntries {
		if len(csvLine) == 0 {
			continue
		}

		split := strings.Split(strings.ReplaceAll(csvLine, "\"", ""), ";")

		date, _ := time.Parse("02.01.2006", split[1])

		if date.Before(now) {
			//fmt.Printf("Skipping date '%s', because it is before '%s'\n", date.Format("2006-01-02"), now.Format("2006-01-02"))
		} else if date.Before(nearestDate) {
			nearestDate = date
		}
	}

	return nearestDate
}

func getGarbageTypes(csvEntries []string, date time.Time) []garbageType {
	var garbageTypes []garbageType

	for _, csvLine := range csvEntries {
		if len(csvLine) == 0 {
			continue
		}

		split := strings.Split(strings.ReplaceAll(csvLine, "\"", ""), ";")

		entry := split[2]
		entryDate, _ := time.Parse("02.01.2006", split[1])

		if entryDate.Equal(date) {
			if strings.Contains(strings.ToLower(entry), "rest") {
				garbageTypes = append(garbageTypes, BLACK)
			}
			if strings.Contains(strings.ToLower(entry), "bio") {
				garbageTypes = append(garbageTypes, BROWN)
			}
			if strings.Contains(strings.ToLower(entry), "tanne") {
				garbageTypes = append(garbageTypes, CHRISTMAS)
			}
			if strings.Contains(strings.ToLower(entry), "papier") {
				garbageTypes = append(garbageTypes, BLUE)
			}
			if strings.Contains(strings.ToLower(entry), "gelb") {
				garbageTypes = append(garbageTypes, YELLOW)
			}
		}
	}

	return garbageTypes
}

type nextResponse struct {
	Date  string        `json:"day_of_collection"`
	Types []garbageType `json:"garbage_types"`
}

type garbageType string

const (
	YELLOW    garbageType = "yellow"
	BLUE                  = "blue"
	BROWN                 = "brown"
	BLACK                 = "black"
	CHRISTMAS             = "christmas"
)
