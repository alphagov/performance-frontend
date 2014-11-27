package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/alext/tablecloth"
	"net/http"
	"os"
	"runtime"
	"sync"
)

var (
	ConfigAPIClient Client
	DataAPIClient   *DataClient
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
		port = getEnvDefault("HTTP_PORT", "8080")
		// databaseName = getEnvDefault("DBNAME", "backdrop")
		// mongoURL     = getEnvDefault("MONGO_URL", "localhost")
		// bearerToken  = getEnvDefault("BEARER_TOKEN", "EMPTY")
		configAPIURL = getEnvDefault("CONFIG_API_URL", "https://stagecraft.preview.performance.service.gov.uk/public/dashboards")
		dataAPIURL   = getEnvDefault("DATA_API_URL", "https://www.preview.performance.service.gov.uk")
		// maxGzipBody  = getEnvDefault("MAX_GZIP_SIZE", "10000000")
		logLevel = getEnvDefault("LOG_LEVEL", "info")
		logger   = newLog(logLevel)
	)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	// handlers.ConfigAPIClient = config.NewClient(configAPIURL, bearerToken)
	// handlers.DataSetStorage = handlers.NewMongoStorage(mongoURL, databaseName)
	// handlers.StatsdClient = handlers.NewStatsDClient("localhost:8125", "datastore.")
	ConfigAPIClient = NewMetaClient(configAPIURL, logger)
	DataAPIClient = NewDataClient(dataAPIURL, logger)

	go serve(":"+port, NewHandler(logger), wg, logger)
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
