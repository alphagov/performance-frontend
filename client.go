package main

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
)

type Dashboards struct {
	Items []Dashboard `json:"Items"`
}

type Dashboard struct {
	Department    Organisation `json:"department"`
	Agency        Organisation `json:"agency,omitempty"`
	DashboardType string       `json:"dashboard-type"`
	Slug          string       `json:"slug"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Modules       []Module     `json:"modules"`
	Published     bool         `json:"published"`
	PageType      string       `json:"page-type"`
	Costs         string       `json:"costs"`
}

type Organisation struct {
	Abbreviation string `json:"abbr"`
	Title        string `json:"title"`
}

type Module struct {
	Info       []string   `json:"info"`
	DataSource DataSource `json:"data-source"`
	Tabs       []Tab      `json:"tabs"`
	Title      string     `json:"title"`
}

type Tab struct {
	Description string     `json:"description"`
	DataSource  DataSource `json:"data-source"`
}

type DataSource struct {
	DataGroup   string      `json:"data-group"`
	DataType    string      `json:"data-type"`
	QueryParams QueryParams `json:"query-params"`
}

type QueryParams struct {
	FilterBy []string `json:"filter_by,omitempty"`
	GroupBy  string   `json:"group_by,omitempty"`
	Collect  []string `json:"collect,omitempty"`
	SortBy   string   `json:"sort_by,omitempty"`
	Duration int      `json:"duration,omitempty"`
	Period   string   `json:"period,omitempty"`
	Limit    int      `json:"limit,omitempty"`
}

// Client defines the interface that we need to talk to the meta data API
type Client interface {
	Fetch(path string) (Dashboard, error)
	FetchDashboards() (Dashboards, error)
}

type defaultClient struct {
	baseURL string
	log     *logrus.Logger
}

// NewMetaClient returns a new Client implementation with sensible defaults.
func NewMetaClient(baseURL string, log *logrus.Logger) Client {
	return &defaultClient{baseURL, log}
}

func (c *defaultClient) Fetch(slug string) (dashboard Dashboard, err error) {
	url := c.baseURL + "?slug=" + slug

	c.log.WithFields(logrus.Fields{
		"url": url,
	}).Debug("Requesting meta data for slug")

	resp, err := NewRequest(url, "empty")

	if err != nil {
		return
	}

	body, err := ReadResponseBody(resp)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &dashboard)
	return
}

func (c *defaultClient) FetchDashboards() (results Dashboards, err error) {
	url := c.baseURL
	resp, err := NewRequest(url, "empty")

	if err != nil {
		return
	}

	body, err := ReadResponseBody(resp)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &results)
	return
}
