package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Controller ...
type Controller struct {
	LoadGenerator *LoadGenerator
	Logger        *log.Logger
}

// GetController ...
func GetController() *Controller {
	c := Controller{}
	c.Logger = log.New(os.Stdout, "[CONTROLLER]", 2)
	return &c
}

func (c *Controller) Log(msg string) {
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// Handler ...
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	ip := GetIP(r)
	if !(strings.HasPrefix(ip, "50.99.77.228") || strings.HasPrefix(ip, "[::1]")) {
		c.Logger.Printf("request came from: %s , rejected\n", ip)
		w.WriteHeader(403)
		return
	}
	// fmt.Println(ip)
	if r.URL.EscapedPath() == "/start" {
		c.Logger.Printf("%s start\n", ip)
		numWorkers, err := strconv.Atoi(r.URL.Query().Get("numWorkers"))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		c.LoadGenerator.GenerateLoad(numWorkers)
		w.WriteHeader(200)
	} else if r.URL.EscapedPath() == "/stop" {
		c.LoadGenerator.Stop()
		c.LoadGenerator = nil
		w.WriteHeader(200)
	} else if r.URL.EscapedPath() == "/action" {

	} else if r.URL.EscapedPath() == "/feedback" {
		c.Logger.Printf("%s feedback requested\n", ip)
		feedback := c.LoadGenerator.GetTestResult()
		b, e := json.Marshal(feedback)
		if e != nil {
			panic(e)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
		c.Logger.Printf("%s feedback sent\n", ip)
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
		c.Logger.Printf("%s prepare\n", ip)
		// fmt.Println("prepare", host, numUsers, alpha, loginRatio, fakeToken, seed, warmUp)
		c.LoadGenerator = GetLoadGenerator("http://" + host)
		c.LoadGenerator.PrepareLoad(numUsers, alpha, loginRatio, fakeToken, int64(seed), warmUp)
		w.WriteHeader(200)
	}
}
