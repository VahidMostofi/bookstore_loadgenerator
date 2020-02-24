package main

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
)

type ResponseHandler func(int, []byte) string

type Results struct {
	StatusCodesHist       map[int]int
	AverageResponseTime   float64
	TotalNumberOfRequests int
	TestDuration          int64
	RequestTypeHist       map[string]int
}

type Request struct {
	Start        int64
	Finish       int64
	Body         string
	Method       string
	URL          string
	Type         string
	Name         string
	AuthRequired bool
	StatusCode   int
	Handle       ResponseHandler
}

type LoadGenerator struct {
	Tokens        map[string]string
	Names         []string
	Books         []string
	RequestsQueue chan *Request
	Results       chan *Request
	Requests      []*Request
	BaseURL       string
	TokensLock    sync.RWMutex
	DoneWorkers   chan int
	Result        chan *Results
}

func GetLoadGenerator(baseUrl string) *LoadGenerator {
	lg := &LoadGenerator{
		Tokens:        make(map[string]string),
		Names:         loadUsernames()[:1000],
		Books:         []string{},
		BaseURL:       baseUrl,
		RequestsQueue: make(chan *Request, 50000),
		Results:       make(chan *Request, 500),
		TokensLock:    sync.RWMutex{},
		Requests:      []*Request{},
		DoneWorkers:   make(chan int, 100),
		Result:        make(chan *Results),
	}

	lg.loadBooks()

	return lg
}

func (lg *LoadGenerator) GenerateLoad(numWokers int) {
	requestsCount := len(lg.RequestsQueue)
	for w := 0; w < numWokers; w++ {
		go lg.worker()
	}

	go func(){
		for r := range lg.Results {
			lg.Requests = append(lg.Requests, r)
			// fmt.Println(len(lg.Requests))
			if len(lg.Requests) == requestsCount {
				close(lg.RequestsQueue)
				break
			}
		}
		lg.GetStats()
	}()
}

func (lg *LoadGenerator) PrepareLoad(numUsers int, alpha int) {

	for u := 0; u < numUsers; u++ {
		lg.RequestsQueue <- lg.GetLoginRequest(lg.Names[u])
	}

	for u := 0; u < alpha*numUsers; u++ {
		r := rand.Intn(50)
		if r == 0 {
			lg.RequestsQueue <- lg.GetLoginRequest(lg.Names[u%numUsers])
		} else {
			if r%2 == 0 {
				bookID := lg.Books[rand.Intn(len(lg.Books))]
				lg.RequestsQueue <- lg.GetGetBookRequest(lg.Names[u%numUsers], bookID)
			} else {
				bookID := lg.Books[rand.Intn(len(lg.Books))]
				lg.RequestsQueue <- lg.GetEditBookRequest(lg.Names[u%numUsers], bookID)
			}

		}
	}
}

func (lg *LoadGenerator) GetStats() {
	var firstRequestTime = lg.Requests[0].Start
	var lastRequestTime int64 = 0

	var results *Results = &Results{}
	var totalResponseTime float64 = 0
	results.StatusCodesHist = make(map[int]int)
	results.RequestTypeHist = make(map[string]int)
	for _, r := range lg.Requests {
		responseTime := r.Finish - r.Start
		totalResponseTime += float64(responseTime)

		if count, ok := results.StatusCodesHist[r.StatusCode]; ok {
			results.StatusCodesHist[r.StatusCode] = count + 1
		} else {
			results.StatusCodesHist[r.StatusCode] = 1
		}

		results.RequestTypeHist[r.Type]++

		if r.Start < firstRequestTime {
			firstRequestTime = r.Start
		}
		if r.Start > lastRequestTime {
			lastRequestTime = r.Start
		}
	}
	results.AverageResponseTime = totalResponseTime / float64(len(lg.Requests))
	results.TestDuration = lastRequestTime - firstRequestTime
	results.TotalNumberOfRequests = len(lg.Requests)

	lg.Result <- results
}

func Log(content string) {

}

func (lg *LoadGenerator) GetToken(name string) string {
	lg.TokensLock.RLock()
	defer lg.TokensLock.RUnlock()
	return lg.Tokens[name]
}

func (lg *LoadGenerator) WriteToken(name, token string) {
	lg.TokensLock.Lock()
	defer lg.TokensLock.Unlock()
	if len(lg.Tokens[name]) > 4 {
		return
	}

	lg.Tokens[name] = token
}

func loadUsernames() []string {
	b, err := ioutil.ReadFile("usernames.txt")
	if err != nil {
		panic(err)
	}
	c := string(b)
	names := strings.Split(c, "\n")
	for i := 0; i < len(names); i++ {
		names[i] = strings.TrimSpace(names[i])
	}
	return names
}

func (lg *LoadGenerator) Stop() {
	panic("Implement THIS!")
}
