package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/3dsinteractive/wrkgo"
)

func main() {

	config := &wrkgo.LoadTestConfig{
		BaseURL:         "http://api.dev.dedepos.com",
		ConcurrentUsers: 1000,
		RunDuration:     time.Second * 10,
		DebugError:      true,
		DebugRequest:    false,
		DebugResponse:   false,
	}

	templates := []*wrkgo.LoadTestTemplate{
		{
			ID:      "0",
			URLPath: "/healthz",
			Timeout: time.Second * 5,
			Method:  "GET",
			Headers: map[string]string{
				"Content-Type":  "application/json; charset=UTF-8",
				"Authorization": "Bearer 2ArxhFy6qSCAhHPP6YFgoMNJCGU",
			},
		},
	}

	// runNumber := 1
	// reqSetupHandler := func(tmpl *wrkgo.LoadTestTemplate, req *wrkgo.LoadTestRequest, prevResp *wrkgo.LoadTestResponse) error {
	// 	runNumberTxt := strconv.Itoa(runNumber)
	// 	input := map[string]interface{}{
	// 		"items": []interface{}{
	// 			map[string]interface{}{
	// 				"inventoryId":  "inv",
	// 				"itemSku":      "sku" + runNumberTxt,
	// 				"categoryGuid": fmt.Sprintf("cat-%d", rand.Int()),
	// 				"lineNumber":   1,
	// 				"price":        rand.Float64(),
	// 				"qty":          rand.Int(),
	// 			},
	// 		},
	// 	}
	// 	req.SetBodyJSON(input)
	// 	runNumber++
	// 	return nil
	// }

	reqSetupHandler := func(tmpl *wrkgo.LoadTestTemplate, req *wrkgo.LoadTestRequest, prevResp *wrkgo.LoadTestResponse) error {
		return nil
	}

	lt := wrkgo.NewLoadTest()
	err := lt.Run(config, templates, reqSetupHandler)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

func RandomMinMax(min int, max int) int {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	return r.Intn(max-min+1) + min
}
