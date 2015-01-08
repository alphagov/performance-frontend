package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/alphagov/performanceplatform-client.go"
	"github.com/go-martini/martini"
	// "github.com/golang/groupcache/singleflight"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/unrolled/render.v1"
)

type DashboardModel struct {
	Dashboard performanceclient.Dashboard
	Data      []performanceclient.BackdropResponse
}

var (
	renderer  = render.New(render.Options{})
	requests  = requestMux(workerPool(5))
	emptyTime time.Time
)

// ReadAPIJob represents a request to the Read API, plus the response
type ReadAPIJob struct {
	// DataSource is the DataSource being queried in the Read API.
	// Responses should be reasonably idempotent and thus we can avoid
	// doing unnecessary work if another request is already in flight.
	DataSource performanceclient.DataSource
	// DataResponse is the response from calling the Read API
	DataResponse *DataResponse
}

type Request struct {
	Job        *ReadAPIJob // changed
	ResultChan chan DataResponse
}

// DataResponse is the response from calling the Read API
type DataResponse struct {
	BackdropResponse *performanceclient.BackdropResponse
	Error            error
}

func workerPool(n int) (chan *ReadAPIJob, chan *ReadAPIJob) {
	jobs := make(chan *ReadAPIJob)
	results := make(chan *ReadAPIJob)

	for i := 0; i < n; i++ {
		go worker(jobs, results)
	}

	return jobs, results
}

func worker(jobs chan *ReadAPIJob, results chan *ReadAPIJob) {
	for job := range jobs {
		dataResponse := fetchDataSource(job.DataSource)
		job.DataResponse = &dataResponse
		results <- job
	}
}

func requestMux(jobs chan *ReadAPIJob, results chan *ReadAPIJob) chan *Request {
	requests := make(chan *Request)

	go func() {
		queues := make(map[string][]*Request)

		for {
			select {
			case request := <-requests:
				job := request.Job
				dataSource := job.DataSource
				URL := DataAPIClient.BuildURL(dataSource.DataGroup, dataSource.DataType, dataSource.QueryParams)
				queues[URL] = append(queues[URL], request)

				if len(queues[URL]) == 1 {
					go func() {
						jobs <- job
					}()
				}

			case job := <-results:
				dataSource := job.DataSource
				URL := DataAPIClient.BuildURL(dataSource.DataGroup, dataSource.DataType, dataSource.QueryParams)

				for _, request := range queues[URL] {
					request.ResultChan <- *job.DataResponse
				}

				delete(queues, URL)
			}
		}
	}()

	return requests
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

func extractModules(responses <-chan DataResponse) (results []performanceclient.BackdropResponse) {
	for r := range responses {
		if r.Error == nil {
			results = append(results, *r.BackdropResponse)
		} else {
			fmt.Println(r.Error.Error())
		}
	}

	return
}

func fetchDataSource(dataSource performanceclient.DataSource) DataResponse {
	// start := time.Now()
	queryParams := dataSource.QueryParams
	br, err := DataAPIClient.Fetch(dataSource.DataGroup, dataSource.DataType, queryParams)
	// log.WithFields(logrus.Fields{
	// 	"url":      DataAPIClient.BuildURL(dataSource.DataGroup, dataSource.DataType, queryParams),
	// 	"duration": time.Since(start).Seconds(),
	// }).Debug("Got response")
	return DataResponse{br, err}
}

func fetchModules(dashboard performanceclient.Dashboard, log *logrus.Logger) (out []ReadAPIJob) {
	for _, m := range dashboard.Modules {
		if len(m.Tabs) > 0 {
			for _, t := range m.Tabs {
				out = append(out, newReadAPIJob(t.DataSource))
			}
		} else {
			out = append(out, newReadAPIJob(m.DataSource))
		}
	}

	return out
}

func merge(jobs []ReadAPIJob) <-chan DataResponse {
	var wg sync.WaitGroup
	out := make(chan DataResponse)

	// Start an output goroutine for each input channel in reports.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c ReadAPIJob) {
		req := Request{&c, make(chan DataResponse)}
		requests <- &req
		out <- (<-req.ResultChan)
		wg.Done()
	}
	wg.Add(len(jobs))
	for _, c := range jobs {
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

func newReadAPIJob(dataSource performanceclient.DataSource) ReadAPIJob {
	// Use a pointer so that we can update the QueryParams
	queryParams := &dataSource.QueryParams

	if queryParams.StartAt != emptyTime && queryParams.EndAt == emptyTime &&
		queryParams.Duration == 0 {
		queryParams.EndAt = time.Now()
	} else if queryParams.StartAt != emptyTime && queryParams.EndAt != emptyTime &&
		queryParams.Duration != 0 {
		queryParams.Duration = 0
	} else if len(queryParams.Period) > 0 &&
		queryParams.Duration == 0 &&
		queryParams.StartAt == emptyTime && queryParams.EndAt == emptyTime {
		queryParams.Duration = periodToDuration(queryParams.Period)
	}

	return ReadAPIJob{dataSource, nil}
}

func periodToDuration(period string) int {
	switch period {
	case "hour":
		return 24
	case "day":
		return 30
	case "week":
		return 9
	case "month":
		return 12
	case "quarter":
		return 24
	default:
		panic(fmt.Sprintf("Unknown period: %q", period))
	}
}
