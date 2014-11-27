package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var server *ghttp.Server
	var client Client

	BeforeEach(func() {
		server = ghttp.NewServer()
		client = NewMetaClient(server.URL(), logrus.New())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Fetch", func() {
		It("Should do a thing", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/"),
					ghttp.RespondWith(http.StatusOK, `{
  "modules": [
    {
      "info":[
      "Data source: Department for Work and Pensions"
      ],
      "value-attribute":"number_of_transactions",
      "description":"",
      "module-type":"kpi",
      "title":"Transactions per year",
      "format":{
        "sigfigs":3,
        "magnitude":true,
        "type":"number"
      },
      "classes":"cols3",
      "slug":"transactions-per-year",
      "data-source":{
        "data-group":"transactional-services",
        "data-type":"summaries",
        "query-params":{
          "sort_by":"_timestamp:descending",
          "filter_by":[
            "service_id:dwp-carers-allowance-new-claims",
            "type:seasonally-adjusted"
          ]
        }
      }
    }
  ],
  "department": {"abbr":"DWP","title":"Department for Work and Pensions"}
}`),
				),
			)
			dashboard, err := client.Fetch("carers-allowance")
			Expect(err).To(BeNil())
			Expect(dashboard).ToNot(BeNil())
			Expect(dashboard.Modules).To(HaveLen(1))
			dataSource := dashboard.Modules[0].DataSource
			Expect(dataSource.DataGroup).To(Equal("transactional-services"))
			Expect(dataSource.DataType).To(Equal("summaries"))
			Expect(dataSource.QueryParams.SortBy).To(Equal("_timestamp:descending"))
		})
	})

})