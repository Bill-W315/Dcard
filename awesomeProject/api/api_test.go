package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

//Note:please clear DB before testing

/*
Admin api good case 1
*/
func TestCreateAdHandler(t *testing.T) {
	requestBody := `
			{"title": "Good case 1",
			"startAt": "2023-12-10T03:00:00.000Z",
			"endAt": "2024-12-31T16:00:00.000Z",
			"conditions":{
				"ageStart": 20,
				"ageEnd": 30,
				"Country:":["TW", "JP"],
				"Platform":["android", "ios"]
				}
			}`
	req, err := http.NewRequest("POST", "/api/v1/ad", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adminAPI)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expected := `{"message":"Successfully added your ad!"}`
	//
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

/*
Admin api good case 2
*/
func TestCreateAdHandler2(t *testing.T) {
	requestBody := `{"title": "Good case 2",
			"startAt": "2023-12-10T03:00:00.000Z",
			"endAt": "2024-12-31T16:00:00.000Z",
			"conditions":{}
            }`
	req, err := http.NewRequest("POST", "/api/v1/ad", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adminAPI)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}

/*
Admin api bad case 1: invalid value,country not in ISO3166
*/
func TestCreateAdHandler3(t *testing.T) {
	requestBody := `{"title": "Bad case 1",
			"startAt": "2023-12-10T03:00:00.000Z",
			"endAt": "2023-12-31T16:00:00.000Z",
			"conditions":{
				"country":["NULL"]
			}
		}`
	req, err := http.NewRequest("POST", "/api/v1/ad", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adminAPI)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

/*
Admin api bad case 2: invalid value, ageStart > ageEnd
*/
func TestCreateAdHandler4(t *testing.T) {
	requestBody := `{"title": "Bad case 2",
			"startAt": "2023-12-10T03:00:00.000Z",
			"endAt": "2023-12-31T16:00:00.000Z",
			"conditions":{
				"country":["TW"],
                "ageStart":31,
                "ageEnd":30
			}
		}`
	req, err := http.NewRequest("POST", "/api/v1/ad", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adminAPI)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

/*
Admin api bad case 3: invalid JSON format: "="
*/
func TestCreateAdHandler5(t *testing.T) {
	requestBody := `{"title": "Bad case 3",
			"startAt"= "2023-12-10T03:00:00.000Z",
			"endAt"= "2023-12-31T16:00:00.000Z",
			"conditions":{
				"country":["TW"],
                "ageStart":31,
                "ageEnd":30
			}
		}`
	req, err := http.NewRequest("POST", "/api/v1/ad", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adminAPI)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

/*
public api good case 1
*/
func TestGetAdsHandler1(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ad", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	publicAPI(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedResponseBody := `{"items":[{"title":"Good case 1","endAt":"2024-12-31T16:00:00Z"},{"title":"Good case 2","endAt":"2024-12-31T16:00:00Z"}]}`
	if strings.TrimSpace(string(responseBody)) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("unexpected response body: got %v want %v", string(responseBody), expectedResponseBody)
	}
}

/*
public api good case 2
*/
func TestGetAdsHandler2(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ad?offset=0&limit=1&platform=android", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	publicAPI(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedResponseBody := `{"items":[{"title":"Good case 1","endAt":"2024-12-31T16:00:00Z"}]}`
	if strings.TrimSpace(string(responseBody)) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("unexpected response body: got %v want %v", string(responseBody), expectedResponseBody)
	}
}

/*
public api good case 3
*/
func TestGetAdsHandler3(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ad?country=TW&country=US", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	publicAPI(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedResponseBody := `{"items":[{"title":"Good case 1","endAt":"2024-12-31T16:00:00Z"},{"title":"Good case 2","endAt":"2024-12-31T16:00:00Z"}]}`
	if strings.TrimSpace(string(responseBody)) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("unexpected response body: got %v want %v", string(responseBody), expectedResponseBody)
	}
}

/*
public api bad case 1: invalid value null
*/
func TestGetAdsHandler4(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ad?country=NULL", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	publicAPI(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedResponseBody := `Invalid param: Country value is invalid`
	if strings.TrimSpace(string(responseBody)) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("unexpected response body: got %v want %v", string(responseBody), expectedResponseBody)
	}
}

/*
public api bad case 2: invalid value age > 100
*/
func TestGetAdsHandler5(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/ad?age=101", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	publicAPI(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	responseBody, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	expectedResponseBody := `Invalid param: Age value is invalid`
	if strings.TrimSpace(string(responseBody)) != strings.TrimSpace(expectedResponseBody) {
		t.Errorf("unexpected response body: got %v want %v", string(responseBody), expectedResponseBody)
	}
}
