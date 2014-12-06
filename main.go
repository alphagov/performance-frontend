package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/alext/tablecloth"
	"github.com/alphagov/performanceplatform-client.go"
	"github.com/shaoshing/train"
	"net/http"
	"os"
	"runtime"
	"sync"
)

var (
	ConfigAPIClient performanceclient.MetaClient
	DataAPIClient   performanceclient.DataClient
)

func main() {
	if os.Getenv("GOMAXPROCS") == "" {
		// Use all available cores if not otherwise specified
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if wd := os.Getenv("GOVUK_APP_ROOT"); wd != "" {
		tablecloth.WorkingDir = wd
	}

	var (
		port         = getEnvDefault("HTTP_PORT", "8080")
		configAPIURL = getEnvDefault("CONFIG_API_URL", "https://stagecraft.preview.performance.service.gov.uk/public/dashboards")
		dataAPIURL   = getEnvDefault("DATA_API_URL", "https://www.preview.performance.service.gov.uk")
		logLevel     = getEnvDefault("LOG_LEVEL", "info")
		logger       = newLog(logLevel)
	)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	ConfigAPIClient = performanceclient.NewMetaClient(configAPIURL, logger)
	DataAPIClient = performanceclient.NewDataClient(dataAPIURL, logger)

	train.ConfigureHttpHandler(nil)
	train.Config.Verbose = true

	go serve(":"+port, NewHandler(logger, train.ServeRequest), wg, logger)
	wg.Wait()
}

func serve(addr string, handler http.Handler, wg *sync.WaitGroup, logger *logrus.Logger) {
	defer wg.Done()
	logger.Fatal(tablecloth.ListenAndServe(addr, handler))
}

func getEnvDefault(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	return val
}

func newLog(level string) *logrus.Logger {
	logger := logrus.New()
	levelConst, err := logrus.ParseLevel(level)

	if err != nil {
		logger.Fatal(err)
	}

	logger.Level = levelConst
	logger.Formatter = &logrus.JSONFormatter{}

	return logger
}
