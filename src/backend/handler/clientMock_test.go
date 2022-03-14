package handler

import "abfallkalender_api/src/backend/client"

type ClientMock struct {
	redirectURL  string
	houseNumbers []string
}

func (mt *ClientMock) GetRedirectUrl(_ string) (string, error) {
	return mt.redirectURL, nil
}

func (mt *ClientMock) GetHouseNumbers(_ string, _ string) (client.HouseNumbers, error) {
	numbers := client.HouseNumbers{}

	for _, number := range mt.houseNumbers {
		numbers = append(numbers, number)
	}

	return numbers, nil
}
