package handler

import "abfallkalender_api/src/backend/client"

type ClientMock struct {
	redirectURL          string
	redirectError        error
	streets              []string
	getStreetsError      error
	houseNumbers         []string
	getHouseNumbersError error
	getICSError          error
	getCSVError          error
	ics                  []byte
	csv                  []byte
}

var controller = Controller{}

// TODO validate parameter
func (mt *ClientMock) GetRedirectUrl(_ string) (string, error) {
	if mt.redirectError != nil {
		return "", mt.redirectError
	}

	return mt.redirectURL, nil
}

// TODO validate parameters
func (mt *ClientMock) GetStreets(_ string) (response client.Streets, err error) {
	if mt.getStreetsError != nil {
		return nil, mt.getStreetsError
	}

	streets := client.Streets{}

	for _, street := range mt.streets {
		streets = append(streets, street)
	}

	return streets, nil
}

// TODO validate parameters
func (mt *ClientMock) GetHouseNumbers(_ string, _ string) (client.HouseNumbers, error) {
	if mt.getHouseNumbersError != nil {
		return nil, mt.getHouseNumbersError
	}

	numbers := client.HouseNumbers{}

	for _, number := range mt.houseNumbers {
		numbers = append(numbers, number)
	}

	return numbers, nil
}

// TODO validate parameters
func (mt *ClientMock) GetICS(_ string, _ string, _ string) ([]byte, error) {
	if mt.getICSError != nil {
		return nil, mt.getICSError
	}

	return mt.ics, nil
}

// TODO validate parameters
func (mt *ClientMock) GetCSV(_ string, _ string, _ string) ([]byte, error) {
	if mt.getCSVError != nil {
		return nil, mt.getCSVError
	}

	return mt.csv, nil
}
