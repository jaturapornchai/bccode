package requestapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Get(url string, authHeader string) ([]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)

	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []interface{}{}, err
	}
	return result["data"].([]interface{}), nil
}
