package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPerformanceGoFrontend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PerformanceGoFrontend Suite")
}
