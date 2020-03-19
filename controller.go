package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Controller ...
type Controller struct {
	LoadGenerator *LoadGenerator
}

// GetController ...
func GetController() *Controller {
	c := Controller{}
	return &c
}

// Handler ...
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.EscapedPath() == "/start" {
		fmt.Println("start")
		numWorkers, err := strconv.Atoi(r.URL.Query().Get("numWorkers"))
		if err != nil {
			w.WriteHeader(400)
		}
		c.LoadGenerator.GenerateLoad(numWorkers)
		w.WriteHeader(200)
	} else if r.URL.EscapedPath() == "/stop" {
		c.LoadGenerator.Stop()
		c.LoadGenerator = nil
		w.WriteHeader(200)
	} else if r.URL.EscapedPath() == "/action" {

	} else if r.URL.EscapedPath() == "/feedback" {
		fmt.Println("feedback requested")
		feedback := c.LoadGenerator.GetTestResult()
		b, e := json.Marshal(feedback)
		if e != nil {
			panic(e)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		fmt.Println("feedback sent")
	} else if r.URL.EscapedPath() == "/prepare" {
		host := r.URL.Query().Get("host")
		if len(host) < 2 {
			w.WriteHeader(400)
			return
		}
		numUsers, err := strconv.Atoi(r.URL.Query().Get("numUsers"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		alpha, err := strconv.Atoi(r.URL.Query().Get("alpha"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		loginRatio, err := strconv.Atoi(r.URL.Query().Get("login"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		seed, err := strconv.Atoi(r.URL.Query().Get("seed"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		fakeToken, err := strconv.ParseBool(r.URL.Query().Get("fakeToken"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		warmUp, err := strconv.ParseFloat(r.URL.Query().Get("warmUp"), 64)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		fmt.Println("prepare", host, numUsers, alpha, loginRatio, fakeToken, seed, warmUp)
		c.LoadGenerator = GetLoadGenerator("http://" + host)
		c.LoadGenerator.PrepareLoad(numUsers, alpha, loginRatio, fakeToken, int64(seed), warmUp)
		w.WriteHeader(200)
	}
}
