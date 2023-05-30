package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ShortDescription *string  `json:"shortDescription" binding:"required"`
	Price            *float64 `json:"price,string" binding:"required"`
}

type Receipt struct {
	Retailer     *string  `json:"retailer" validate:"presence"`
	PurchaseDate *string  `json:"purchaseDate"`
	PurchaseTime *string  `json:"purchaseTime"`
	Total        *float64 `json:"total,string"`
	Items        *[]*Item
}

// map of the receipt id to points
var pointsMap map[string]float64 = make(map[string]float64)

type Rid struct {
	Id string `json:"id"`
}

type Rpoints struct {
	Points float64 `json:"points,string"`
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// function that handles the get method to get ponits
func getPoints(w http.ResponseWriter, id string) {

	var points Rpoints
	var ok bool
	points.Points, ok = pointsMap[id]
	// checks if id is in map
	if ok {
		w.Header().Set("Content-Type", "application/json")
		//if yes shwo points
		encoder := json.NewEncoder(w)
		err := encoder.Encode(points)
		if err != nil {
			http.Error(w, err.Error(), 400)
		}

	} else {
		//else error for no such id
		http.Error(w, "No receipt found for that id", 404)
		return
	}

}

// method that calculates the ponts based on the receipt
func calculatePoints(receipt *Receipt) (float64, error) {

	var points = 0.0
	var err error

	//checks that all valid fields are present
	if receipt.Retailer == nil || receipt.PurchaseDate == nil || receipt.PurchaseTime == nil || receipt.Total == nil || receipt.Items == nil {
		return -1, errors.New("invalid json")
	}

	//splits time into hour and minute
	stime := strings.Split(*receipt.PurchaseTime, ":")
	var ptime [2]int

	//checks that time is valid
	if len(stime) == 2 {

		ptime[0], err = strconv.Atoi(stime[0])
		if err != nil || ptime[0] < 0 || ptime[0] > 23 {
			return -1, errors.New("invalid json")
		}

		ptime[1], err = strconv.Atoi(stime[1])
		if err != nil || ptime[1] < 0 || ptime[1] > 59 {
			return -1, errors.New("invalid json")
		}

	} else {
		return -1, errors.New("invalid json")
	}

	//checks if date is valid
	layout := "2006-01-02"
	pDate, err := time.Parse(layout, *receipt.PurchaseDate)
	if err != nil {
		return -1, errors.New("invalid json")
	}

	//checks if each item in Items has all the required fields
	for i := 0; i < len(*receipt.Items); i++ {
		if (*receipt.Items)[i].ShortDescription == nil || (*receipt.Items)[i].Price == nil {
			return -1, errors.New("invalid json")
		}
	}

	//removes all nonalphanumeric characters from retailer
	tempRetailer := nonAlphanumericRegex.ReplaceAllString(*receipt.Retailer, "")

	points = points + float64(len(tempRetailer))

	if *receipt.Total-math.Floor(*receipt.Total) == 0 {
		points = points + 50
	}

	if *receipt.Total/0.25-math.Floor(*receipt.Total/0.25) == 0 {
		points = points + 25
	}

	points = points + (float64(len(*receipt.Items)/2))*5

	for i := 0; i < len(*receipt.Items); i++ {
		if len(strings.TrimSpace(*(*receipt.Items)[i].ShortDescription))%3 == 0 {
			points = points + math.Ceil(*(*receipt.Items)[i].Price*0.2)
		}
	}

	if (ptime[0] >= 14 && ptime[0] < 16) || (ptime[0] == 16 && ptime[1] == 0) {
		points = points + 10
	}

	if pDate.Day()%2 == 1 {
		points = points + 6
	}

	return points, err
}

// handles the post method
func processReciept(w http.ResponseWriter, r *http.Request) {

	//reads the body of the post
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//stores the json into the struct
	var receipt *Receipt
	err = json.Unmarshal(body, &receipt)
	if err != nil {
		http.Error(w, "The receipt is invalid", 400)
		return
	}

	points, err := calculatePoints(receipt)
	if err != nil {
		http.Error(w, "The receipt is invalid", 400)
		return
	}

	var id Rid
	id.Id = generateId()

	pointsMap[id.Id] = points
	w.Header().Set("Content-Type", "application/json")

	//returns newly generated id
	encoder := json.NewEncoder(w)
	err = encoder.Encode(id)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
}

// generates a unique id
func generateId() string {
	id := uuid.NewString()
	for {
		_, ok := pointsMap[id]
		if ok {
			id = uuid.NewString()
		} else {
			break
		}
	}
	return id
}

// handles all requests at /receipts endpoint
func handleRequests(w http.ResponseWriter, r *http.Request) {

	urlSplit := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch r.Method {
	case "GET":
		//checks for correct get enpoint
		if len(urlSplit) == 3 && urlSplit[2] == "points" && urlSplit[0] == "receipts" {
			getPoints(w, urlSplit[1])
		} else {
			http.Error(w, "Bad request", 400)
		}

	case "POST":
		//checks for correct post endpoint
		if len(urlSplit) == 2 && urlSplit[1] == "process" && urlSplit[0] == "receipts" {
			processReciept(w, r)
		} else {
			http.Error(w, "Bad request", 400)
		}
	default:
		http.Error(w, "Bad request", 400)
	}
}

func main() {
	http.HandleFunc("/", handleRequests)
	http.ListenAndServe(":8080", nil)
}
