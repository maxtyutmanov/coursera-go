package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"encoding/xml"
	"io/ioutil"
	"sort"
	"strings"
	"net/http/httptest"
	"testing"
	"time"
)

type DsPerson struct {
	XMLName xml.Name `xml:"row"`
	ID int `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName string `xml:"last_name"`
	Age int `xml:"age"`
	About string `xml:"about"`
	Gender string `xml:"gender"`
}

func (p *DsPerson) ToUser() User {
	return User {
		Id: p.ID,
		Name: p.FirstName + " " + p.LastName,
		Age: p.Age,
		Gender: p.Gender,
		About: p.About,
	}
}

type RawDataset struct {
	XMLName xml.Name `xml:"root"`
	Persons []DsPerson `xml:"row"`
}

type sortFunc func(i, j int) bool

func loadUsers() []User {
	dsBytes, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err)
	}
	rds := RawDataset{}
	xml.Unmarshal(dsBytes, &rds)

	users := make([]User, 0, len(rds.Persons))

	for _, dsPerson := range rds.Persons {
		users = append(users, dsPerson.ToUser())
	}

	return users
}

func sortUsers(users []User, orderField string, orderType int) error {
	var less sortFunc
	switch orderField {
	case "Id":
		less = func (i, j int) bool {
			if (orderType == OrderByAsc) {
				return users[i].Id < users[j].Id
			} else if (orderType == OrderByDesc) {
				return users[i].Id > users[j].Id
			} else {
				return i < j
			}
		}
	case "Age":
		less = func (i, j int) bool {
			if (orderType == OrderByAsc) {
				return users[i].Age < users[j].Age
			} else if (orderType == OrderByDesc) {
				return users[i].Age > users[j].Age
			} else {
				return i < j
			}
		}
	case "", "Name":
		less = func (i, j int) bool {
			if (orderType == OrderByAsc) {
				return users[i].Name < users[j].Name
			} else if (orderType == OrderByDesc) {
				return users[i].Name > users[j].Name
			} else {
				return i < j
			}
		}
	default:
		return fmt.Errorf(ErrorBadOrderField)
	}

	sort.SliceStable(users, less)
	return nil
}

func filterUsers(users []User, query string) []User {
	if query == "" {
		return users
	}

	filtered := make([]User, 0, 0)

	for _, user := range users {
		if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
			filtered = append(filtered, user)
		}
	}

	return filtered
}

func pageUsers(users []User, offset int, limit int) []User {
	left := offset
	right := offset + limit
	
	if left >= len(users) {
		return make([]User, 0, 0)
	}

	if right > len(users) {
		right = len(users)
	}

	return users[left:right]
}

func writeBadRequest(w http.ResponseWriter, message string) {
	errResp := SearchErrorResponse{
		Error: message,
	}

	errRespSerialized, _ := json.Marshal(errResp)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(errRespSerialized)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))
	accessToken := r.Header.Get("AccessToken")

	if accessToken == "INVALID" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if query == "QUERY_THAT_BREAKS_EVERYTHING" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if query == "I_JUST_DONT_LIKE_THAT_QUERY" {
		writeBadRequest(w, "GoToHellWithYourQuery")
		return
	}

	if query == "HEAVY_QUERY_THAT_WILL_TIME_OUT" {
		time.Sleep(time.Second * 2)
	}

	if query == "UNKNOWN_ERROR_QUERY" {
		return
	}

	if query == "GIVE_ME_INVALID_JSON" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("abracodabra"))
		return
	}
	
	if query == "GIVE_ME_INVALID_JSON_BUT_SUCCESS" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("abracodabra"))
		return
	}

	users := loadUsers()
	users = filterUsers(users, query)
	err := sortUsers(users, orderField, orderBy)

	if err != nil {
		writeBadRequest(w, "ErrorBadOrderField")
		return
	}

	users = pageUsers(users, offset, limit)
	usersStr, _ := json.Marshal(users)

	w.WriteHeader(http.StatusOK)
	w.Write(usersStr)
}

func createClientWithoutServer() SearchClient {
	client := SearchClient{
		URL: "",
		AccessToken: "VALID",
	}
	return client
}

func createServerAndClient(accessToken string) SearchClient {
	if accessToken == "" {
		accessToken = "VALID"
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	
	client := SearchClient{
		URL: ts.URL,
		AccessToken: accessToken,
	}

	return client
}

func TestNegativeLimit(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: -1,
	})

	if resp != nil || err == nil {
		t.Fail()
	}
}

func TestLimitLargerThanMaximum(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 30,
		Offset: 0,
	})

	if err != nil || len(resp.Users) != 25 {
		t.Fail()
	}
}

func TestNegativeOffset(t *testing.T) {
	client := createServerAndClient("")

	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: -1,
	})

	if resp != nil || err == nil {
		t.Fail()
	}
}

func TestTimeout(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "HEAVY_QUERY_THAT_WILL_TIME_OUT",
	})

	if resp != nil || err == nil {
		t.Fail()
	}
}

func TestUnknownSendError(t *testing.T) {
	client := createClientWithoutServer()
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "UNKNOWN_ERROR_QUERY",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "unknown error") {
		t.Fail()
	}
}

func TestBadAccessToken(t *testing.T) {
	client := createServerAndClient("INVALID")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "sample query",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "Bad AccessToken") {
		t.Fail()
	}
}

func TestInternalServerError(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "QUERY_THAT_BREAKS_EVERYTHING",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "SearchServer fatal error") {
		t.Fail()
	}
}

func TestInvalidJsonInBadRequestResponse(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "GIVE_ME_INVALID_JSON",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "cant unpack error json") {
		t.Fail()
	}
}

func TestOrderByWrongField(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "sample query",
		OrderField: "About",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "OrderFeld About invalid") {
		t.Fail()
	}
}

func TestOkStatusButInvalidJson(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
		Query: "GIVE_ME_INVALID_JSON_BUT_SUCCESS",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "cant unpack result json") {
		t.Fail()
	}
}

func TestNextPageAvailable(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 0,
	})

	if resp == nil || err != nil || !resp.NextPage {
		t.Fail()
	}

	if len(resp.Users) != 10 {
		t.Fail()
	}
}

func TestNextPageIsNotAvailable(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 30,
	})

	if resp == nil || err != nil || resp.NextPage {
		t.Fail()
	}

	if len(resp.Users) != 5 {
		t.Fail()
	}
}

func TestUnknownBadRequest(t *testing.T) {
	client := createServerAndClient("")
	
	resp, err := client.FindUsers(SearchRequest{
		Limit: 10,
		Offset: 30,
		Query: "I_JUST_DONT_LIKE_THAT_QUERY",
	})

	if resp != nil || err == nil || !strings.Contains(err.Error(), "unknown bad request error") {
		t.Fail()
	}
}