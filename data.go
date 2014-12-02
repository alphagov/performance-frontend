package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-querystring/query"
)

type DataClient struct {
	URL string

	log *logrus.Logger
}

func NewDataClient(url string, logger *logrus.Logger) *DataClient {
	return &DataClient{
		URL: url,
		log: logger,
	}
}

type BackdropResponse struct {
	Data    json.RawMessage `json:"data"`
	Warning string          `json:"warning,omitempty"`
	Status  string          `json:"status,omitempty"`
	Message string          `json:"message,omitempty"`
}

func (client *DataClient) BuildURL(dataGroup, dataType string, dataQuery QueryParams) string {
	url := fmt.Sprintf("%s/data/%s/%s", client.URL, dataGroup, dataType)

	values, _ := query.Values(dataQuery)
	queryParameters := values.Encode()

	if len(queryParameters) > 1 {
		url += "?" + queryParameters
	}

	return url
}

func (client *DataClient) Fetch(dataGroup, dataType string, dataQuery QueryParams) (*BackdropResponse, error) {
	url := client.BuildURL(dataGroup, dataType, dataQuery)

	client.log.WithFields(logrus.Fields{
		"url": url,
	}).Debug("Requesting performance data for slug")

	backdropResponse, err := NewRequest(url, "EMPTY")
	if err != nil {
		return nil, err
	}

	backdropBody, err := ReadResponseBody(backdropResponse)
	if err != nil {
		return nil, err
	}

	backdrop, err := ParseBackdropResponse([]byte(backdropBody))
	if err != nil {
		return nil, err
	}

	return backdrop, nil
}

func ParseBackdropResponse(response []byte) (*BackdropResponse, error) {
	backdropResponse := &BackdropResponse{}
	if err := json.Unmarshal(response, &backdropResponse); err != nil {
		return nil, err
	}

	if backdropResponse.Status == "error" {
		return nil, errors.New(backdropResponse.Message)
	}

	return backdropResponse, nil
}
