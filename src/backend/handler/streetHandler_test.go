package handler

import (
	"abfallkalender_api/src/backend/client"
	"testing"
)

type MyTest struct {
	called bool
}

func (mt *MyTest) GetRedirectUrl(url string) (string, error) {
	mt.called = true
	return "", nil
}

func (mt *MyTest) GetHouseNumbers(url string, streetName string) (client.HouseNumbers, error) {
	return nil, nil
}

func Test(t *testing.T) {
	test := MyTest{}

	controller := Controller{
		Client: &test,
	}

	// TODO use request recorder
	// TODO use test request

	controller.GetStreet()
}
