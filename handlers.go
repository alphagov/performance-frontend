package main

import (
	// "fmt"
	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/unrolled/render.v1"
)

type DashboardModel struct {
	Dashboard Dashboard
	Data      []BackdropResponse
}

var (
	renderer = render.New(render.Options{})
)

type DataResponse struct {
	BackdropResponse *BackdropResponse
	Error            error
}

func NewHandler(logger *logrus.Logger) http.Handler {
	m := martini.Classic()
	m.Map(logger)
	m.Get("/performance", HomepageHandler)
	m.Get("/performance/**", ProcessRequestHandler)
	return m
}

func HomepageHandler(w http.ResponseWriter, r *http.Request, log *logrus.Logger) {
	dashboards, err := ConfigAPIClient.FetchDashboards()
	// render dashboards
	if err != nil {
		log.Error(err)
		renderError(w, err)
		return
	}

	renderer.HTML(w, http.StatusOK, "home", dashboards)
}

func ProcessRequestHandler(w http.ResponseWriter, r *http.Request, log *logrus.Logger) {
	path := r.URL.Path
	slug := strings.Replace(path, "/performance/", "", -1)
	dashboard, err := ConfigAPIClient.Fetch(slug)

	if err != nil {
		log.Error(err)
		renderError(w, err)
		return
	}

	responses := merge(fetchModules(dashboard, log))

	modules := extractModules(responses)

	renderer.HTML(w, http.StatusOK, "dashboard", DashboardModel{dashboard, modules})
}

func renderError(w http.ResponseWriter, err error) {
	renderer.HTML(w, http.StatusInternalServerError, "error", err)
}

func extractModules(responses <-chan DataResponse) (results []BackdropResponse) {
	for r := range responses {
		if r.Error == nil {
			results = append(results, *r.BackdropResponse)
		}
	}

	return
}

func fetchModules(dashboard Dashboard, log *logrus.Logger) chan DataResponse {
	out := make(chan DataResponse)
	fetchDataSource := func(dataSource DataSource) DataResponse {
		start := time.Now()
		queryParams := dataSource.QueryParams
		br, err := DataAPIClient.Fetch(dataSource.DataGroup, dataSource.DataType, queryParams)
		log.WithFields(logrus.Fields{
			"url":      DataAPIClient.BuildURL(dataSource.DataGroup, dataSource.DataType, queryParams),
			"duration": time.Since(start).Seconds(),
		}).Debug("Got response")
		return DataResponse{br, err}
	}

	go func() {
		defer close(out)
		for _, m := range dashboard.Modules {
			if len(m.Tabs) > 0 {
				for _, t := range m.Tabs {
					out <- fetchDataSource(t.DataSource)
				}
			} else {
				out <- fetchDataSource(m.DataSource)
			}
		}
	}()
	return out
}

func merge(reports ...chan DataResponse) <-chan DataResponse {
	var wg sync.WaitGroup
	out := make(chan DataResponse)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan DataResponse) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(reports))
	for _, c := range reports {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
