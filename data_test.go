package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Data", func() {
	Describe("ParseBackdropResponse", func() {
		It("returns an error when it can't parse the string", func() {
			backdrop, err := ParseBackdropResponse([]byte(""))

			Expect(err).ToNot(BeNil())
			Expect(backdrop).To(BeNil())
		})

		It("parses a Backdrop API response into a Backdrop struct", func() {
			exampleResponse := strings.TrimSpace(`
{
  "data": [
    {
      "_count": 1.0,
      "searchKeyword": "2014 level thresholds",
      "searchUniques": 4,
      "searchUniques:sum": 12.0
    }
  ],
  "warning": "Warning: This data-set is unpublished. Data may be subject to change or be inaccurate."
}`)

			backdrop, err := ParseBackdropResponse([]byte(exampleResponse))
			Expect(err).To(BeNil())
			Expect(backdrop).ToNot(BeNil())
			Expect(backdrop.Warning).To(Equal("Warning: This data-set is unpublished. Data may be subject to change or be inaccurate."))

			var data []interface{}
			err = json.Unmarshal(backdrop.Data, &data)
			Expect(err).To(BeNil())
			Expect(len(data)).To(Equal(1))
		})

		It("parses a Backdrop API response and errors appropriately", func() {
			exampleResponse := strings.TrimSpace(`
{
  "status": "error",
  "message": "Ahh an error happened"
}`)

			backdrop, err := ParseBackdropResponse([]byte(exampleResponse))
			Expect(err).To(MatchError("Ahh an error happened"))
			Expect(backdrop).To(BeNil())
		})
	})

	Describe("Client", func() {
		var server *ghttp.Server
		var client *DataClient

		BeforeEach(func() {
			server = ghttp.NewServer()
			client = NewDataClient(server.URL(), logrus.New())
		})

		AfterEach(func() {
			server.Close()
		})

		Describe("BuildURL", func() {
			It("Should return a url with no parameters", func() {
				client := NewDataClient("http://perf", nil)
				Expect(client.BuildURL("govuk-info", "page-statistics", QueryParams{})).
					To(Equal("http://perf/data/govuk-info/page-statistics"))
			})

			It("Should add filter by parameters", func() {
				client := NewDataClient("http://perf", nil)
				Expect(client.BuildURL("govuk-info", "page-statistics", QueryParams{
					FilterBy: []string{"pagePath:/bank-holidays", "bar:foo"},
				})).To(Equal("http://perf/data/govuk-info/page-statistics?filter_by=pagePath%3A%2Fbank-holidays&filter_by=bar%3Afoo"))
			})
		})

		Describe("Fetch", func() {
			It("Should do a thing", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/data/govuk-info/page-statistics"),
						ghttp.RespondWith(http.StatusOK, `
{
  "data": [
    {
      "_count": 1.0,
      "searchKeyword": "2014 level thresholds",
      "searchUniques": 4,
      "searchUniques:sum": 12.0
    }
  ],
  "warning": "Warning: This data-set is unpublished. Data may be subject to change or be inaccurate."
}`),
					),
				)
				response, err := client.Fetch("govuk-info", "page-statistics", QueryParams{})
				Expect(err).To(BeNil())
				Expect(response).ToNot(BeNil())

				var data []interface{}
				err = json.Unmarshal(response.Data, &data)
				Expect(err).To(BeNil())
				Expect(len(data)).To(Equal(1))
			})
		})
	})

})
