package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var reqID = 0

func (lg *LoadGenerator) GetLoginRequest(name string) *Request {
	r := Request{
		Method:       "POST",
		URL:          "/auth/login",
		Body:         fmt.Sprintf(`{"email":"%s","password":"123456789"}`, name+"@gmail.com"),
		Handle:       HandleLoginResponse,
		Name:         name,
		Type:         "login",
		AuthRequired: false,
		ID:           reqID,
	}
	reqID++
	return &r
}

func (lg *LoadGenerator) GetGetBookRequest(name string, bookId string) *Request {
	r := Request{
		Method:       "GET",
		URL:          "/books/" + bookId,
		Body:         "",
		Handle:       nil,
		Name:         name,
		Type:         "getbook",
		AuthRequired: true,
		ID:           reqID,
	}
	reqID++
	return &r
}

func (lg *LoadGenerator) GetEditBookRequest(name string, bookId string) *Request {
	r := Request{
		Method:       "PUT",
		URL:          "/books/" + bookId,
		Body:         fmt.Sprintf(`{"description":"%s","pages":%d}`, "some new description", 200+rand.Intn(200)),
		Handle:       nil,
		Name:         name,
		Type:         "editbook",
		AuthRequired: true,
		ID:           reqID,
	}
	reqID++
	return &r
}

func HandleLoginResponse(statusCode int, resp []byte) string {
	type Temp struct {
		Token string `json:"token"`
	}
	var temp Temp
	json.Unmarshal(resp, &temp)
	return temp.Token
}

func (lg *LoadGenerator) loadBooks() {
	url := lg.BaseURL + "/books/"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJfaWQiOiI1ZTUxNTRmN2MzZGE0OTAwMTE3M2U3M2MiLCJpYXQiOjE1ODIzODg1NzZ9.D-iHOtrbJznER5lNc8Ta_lQmcJflqgqmmZdQXvdhMXo")

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	type temp struct {
		ID string `json:"_id"`
	}
	var arr []temp
	json.Unmarshal(body, &arr)
	for _, t := range arr {
		lg.Books = append(lg.Books, t.ID)
	}
}

func (lg *LoadGenerator) MakeRequest(r *Request, debug bool) (*Request, bool) {
	url := lg.BaseURL + r.URL
	method := r.Method
	client := &http.Client{}

	var payload *strings.Reader
	var req *http.Request
	var err error
	if method == "POST" || method == "PUT" {
		payload = strings.NewReader(r.Body)
		req, err = http.NewRequest(method, url, payload)
	} else if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	var token string
	if r.AuthRequired {
		token = lg.GetToken(r.Name)
		if len(token) < 3 {
			if debug {
				fmt.Println(method, url, r.Body, r.AuthRequired, r.Name, "postponed")
			}
			return r, false
		}
		req.Header.Add("Authorization", "Bearer "+token)
	}

	if debug {
		fmt.Println("ID", r.ID, r.Type)
	}
	r.Start = time.Now().UnixNano() / 1e6
	res, err := client.Do(req)
	r.Finish = time.Now().UnixNano() / 1e6
	if debug {
		fmt.Println("ID", -(r.ID))
	}

	if err != nil {
		panic(err)
	}
	if debug {
		fmt.Println(r.Type, method, url, r.Body, r.AuthRequired, res.StatusCode, r.Name)
		if res.StatusCode == 400 {
			b, _ := ioutil.ReadAll(res.Body)
			fmt.Println(string(b))
		} else if res.StatusCode == 401 {
			b, _ := ioutil.ReadAll(res.Body)
			fmt.Println(r.Name, "this is token", len(token), token, string(b))
		} else if res.StatusCode != 200 {
			b, _ := ioutil.ReadAll(res.Body)
			fmt.Println(res.StatusCode, string(b))
		}
	}
	r.StatusCode = res.StatusCode
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Log(err.Error())
	}
	if r.Handle != nil {
		if r.Type == "login" {
			token := r.Handle(res.StatusCode, body)
			if debug {
				fmt.Println("got token", r.Name, len(token))
			}
			lg.WriteToken(r.Name, token)
		}
	}
	return r, true
}
