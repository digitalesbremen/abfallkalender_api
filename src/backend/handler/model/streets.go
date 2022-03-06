package model

type StreetsDto struct {
	Embedded struct {
		Streets []StreetDto `json:"streets"`
	} `json:"_embedded"`
}

type StreetDto struct {
	Name  string `json:"name"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}
