package handler

import "abfallkalender_api/src/backend/client"

type ClientMock struct {
	redirectURL          string
	redirectError        error
	houseNumbers         []string
	getHouseNumbersError error
}

func (mt *ClientMock) GetRedirectUrl(_ string) (string, error) {
	if mt.redirectError != nil {
		return "", mt.redirectError
	}

	return mt.redirectURL, nil
}

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
