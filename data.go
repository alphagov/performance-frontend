package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-querystring/query"
)

type Query struct {
	FilterBy []string `url:"filter_by,omitempty"`
	GroupBy  string   `url:"group_by,omitempty"`
	Collect  []string `url:"collect,omitempty"`
	SortBy   string   `url:"sort_by,omitempty"`

	Duration int    `url:"duration,omitempty"`
	Period   string `url:"period,omitempty"`

	Limit int `url:"limit,omitempty"`

	StartAt time.Time `url:"start_at,omitempty"`
	EndAt   time.Time `url:"end_at,omitempty"`
}

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

func (client *DataClient) BuildURL(dataGroup, dataType string, dataQuery Query) string {
	url := fmt.Sprintf("%s/data/%s/%s", client.URL, dataGroup, dataType)

	values, _ := query.Values(dataQuery)
	queryParameters := values.Encode()

	if len(queryParameters) > 1 {
		url += "?" + queryParameters
	}

	return url
}

func (client *DataClient) Fetch(dataGroup, dataType string, dataQuery Query) (*BackdropResponse, error) {
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
