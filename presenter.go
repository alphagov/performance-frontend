package main

import (
	"math"
)

import (
	"github.com/alphagov/performanceplatform-client.go"
)

func partition(list []performanceclient.Dashboard, fn func(performanceclient.Dashboard, int) bool) (yes []performanceclient.Dashboard, no []performanceclient.Dashboard) {
	for i, elem := range list {
		if fn(elem, i) {
			yes = append(yes, elem)
		} else {
			no = append(no, elem)
		}
	}
	return
}

func homePage(services, serviceGroups, highVolumeServices, contentDashboards []performanceclient.Dashboard) map[string]interface{} {
	midPoint := int(math.Ceil(float64((len(services) + len(serviceGroups))) / 2))
	firstServices, secondServices := partition(services, func(d performanceclient.Dashboard, i int) bool {
		return (i + 1) <= midPoint
	})

	midPoint = int(math.Ceil(float64((len(contentDashboards))) / 2))
	firstContent, secondContent := partition(contentDashboards, func(d performanceclient.Dashboard, i int) bool {
		return (i + 1) <= midPoint
	})

	return map[string]interface{}{
		"assetPath":          "/assets/",
		"serviceCount":       len(services),
		"firstServices":      firstServices,
		"secondServices":     secondServices,
		"serviceGroups":      serviceGroups,
		"highVolumeServices": highVolumeServices,
		"firstContent":       firstContent,
		"secondContent":      secondContent,
	}
}

func dashboardPage(dashboard performanceclient.Dashboard, modules interface{}) map[string]interface{} {
	return map[string]interface{}{
		"assetPath": "/assets/",
		"dashboard": dashboard,
	}
}
