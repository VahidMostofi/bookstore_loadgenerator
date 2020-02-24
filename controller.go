package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Controller struct {
	LoadGenerator *LoadGenerator
}

func GetController() *Controller {
	c := Controller{}
	return &c
}

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
		fmt.Println("feedback")
		r := <-c.LoadGenerator.Result
		b, e := json.Marshal(r)
		if e != nil {
			panic(e)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
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
		fmt.Println("prepare", host, numUsers, alpha)
		c.LoadGenerator = GetLoadGenerator("http://" + host)
		c.LoadGenerator.PrepareLoad(numUsers, alpha)
		w.WriteHeader(200)
	}
}
