package datagateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type responseBody struct {
	Results []struct {
		Rows []map[string]interface{} `json:"rows"`
	} `json:"results"`
}

func RunQuery(dataGatewayUrl string, query string) ([]map[string]interface{}, error) {

	payload := map[string]interface{}{
		"sql": query,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payload: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, dataGatewayUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("got status code: %v from data gateway : %v", resp.StatusCode, string(bodyBytes))
	}

	response := responseBody{}
	err = json.Unmarshal([]byte(bodyBytes), &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return response.Results[0].Rows, nil

}
