package api

import (
    "encoding/json"
    "github.com/gorilla/mux"
    _ "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    "log"
    "net/http"
    "time"
)

// Admin api request
type Ad struct {
    UUID       string      `json:"-"`
    Title      string      `json:"title"`
    StartAt    time.Time   `json:"startAt"`
    EndAt      time.Time   `json:"endAt"`
    Conditions AdCondition `json:"conditions"`
}

// Admin api request
type AdCondition struct {
    AgeStart  int      `json:"ageStart"`
    AgeEnd    int      `json:"ageEnd"`
    Gender    []string `json:"Gender"`
    Countries []string `json:"Country"`
    Platforms []string `json:"Platform"`
}

// Public api request
type SearchCondition struct {
    Offset   int      `json:"-"`
    Limit    int      `json:"-"`
    Age      []string `json:"age"`
    Gender   []string `json:"gender"`
    Country  []string `json:"country"`
    Platform []string `json:"platform"`
}

// Public api response
type SearchResult struct {
    Title string    `json:"title"`
    EndAt time.Time `json:"endAt"`
}

func Main() {
    //Init DB,Redis
    setConnections()

    //Register api handler
    r := mux.NewRouter()
    r.HandleFunc("/api/v1/ad", adminAPI).Methods("POST")
    r.HandleFunc("/api/v1/ad", publicAPI).Methods("GET")

    //Clear cache
    clearSearchHistory()

    //Start Server
    println("Server started on port 8080 ", getNowTime().String())
    log.Fatal(http.ListenAndServe(":8080", r))
}

func adminAPI(w http.ResponseWriter, r *http.Request) {

    //For api test
    //setConnections()

    //Http method check
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    //JSON format check
    var ad Ad
    err := json.NewDecoder(r.Body).Decode(&ad)
    if err != nil {
        http.Error(w, "Failed to decode JSON request body "+err.Error(), http.StatusBadRequest)
        return
    }

    //JSON value check
    err = validateAd(ad)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    //Insert Ad into storage
    err = saveAd(ad)
    if err != nil {
        http.Error(w, "Database error "+err.Error(), http.StatusInternalServerError)
        return
    }

    //Clear cache
    if ad.StartAt.Before(getNowTime()) && ad.EndAt.After(getNowTime()) {
        clearSearchHistory()
    }

    //Response body
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Successfully added your ad!"})
}

func publicAPI(w http.ResponseWriter, r *http.Request) {

    //For api test
    //setConnections()

    //Http method check
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    //Validate URL param and assign default value
    condition, err := validateSearchParamAndAssignDefaultVal(r)
    if err != nil {
        http.Error(w, "Invalid param: "+err.Error(), http.StatusBadRequest)
        return
    }

    //Find Ad matches search conditions
    ads, err := getAdsByConditions(condition)
    if err != nil {
        http.Error(w, "Invalid condition: "+err.Error(), http.StatusBadRequest)
        return
    }

    //Response body
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"items": ads})
}
