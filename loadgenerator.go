package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"

	"github.com/montanaflynn/stats"
)

// ResponseHandler handles the result
type ResponseHandler func(int, []byte) string

// TestResult Aggregates information about one workload test
type TestResult struct {
	TotalNumberOfRequests int
	TestDuration          int64
	Requests              map[string]*RequestResult
	ConcurrencyInfo       *ConcurrencyInfo
	WorkerCount           int
	Alpha                 int
	UsersCount            int
	Unit                  string
}

// RequestResult ...
type RequestResult struct {
	AverageResponseTime float64
	Percentile95        float64
	Percentile99        float64
	Count               int
	StatusCodesHist     map[int]int
	responseTimes       []float64
}

// ConcurrencyInfo ...
type ConcurrencyInfo struct {
	MaxConcurrentRequestsInOneUnit     float64
	Percentile95                       float64
	Percentile99                       float64
	AverageConcurrentRequestsInOneUnit float64
	HowManyUnitsIsASecond              float64
}

// Request ...
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

// LoadGenerator ...
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
	result        chan *TestResult

	NumUsers   int
	NumWorkers int
	Alpha      int
}

// GetTestResult ....
func (l *LoadGenerator) GetTestResult() *TestResult {
	r := <-l.result
	go func() { l.result <- r }()
	return r
}

// GetLoadGenerator ...
func GetLoadGenerator(baseURL string) *LoadGenerator {
	lg := &LoadGenerator{
		Tokens:        make(map[string]string),
		Names:         loadUsernames()[:1000],
		Books:         []string{},
		BaseURL:       baseURL,
		RequestsQueue: make(chan *Request, 50000),
		Results:       make(chan *Request, 500),
		TokensLock:    sync.RWMutex{},
		Requests:      []*Request{},
		DoneWorkers:   make(chan int, 100),
		result:        make(chan *TestResult),
	}

	lg.loadBooks()

	return lg
}

// GenerateLoad ...
func (l *LoadGenerator) GenerateLoad(numWokers int) {
	l.NumWorkers = numWokers
	requestsCount := len(l.RequestsQueue)
	for w := 0; w < numWokers; w++ {
		go l.worker()
	}

	go func() {
		for r := range l.Results {
			l.Requests = append(l.Requests, r)
			// fmt.Println(len(lg.Requests))
			if len(l.Requests) == requestsCount {
				close(l.RequestsQueue)
				break
			}
		}
		l.GetStats()
	}()
}

// PrepareLoad ...
func (l *LoadGenerator) PrepareLoad(numUsers int, alpha int) {
	rand.Seed(8)
	l.NumUsers = numUsers
	l.Alpha = alpha
	for u := 0; u < numUsers; u++ {
		l.RequestsQueue <- l.GetLoginRequest(l.Names[u])
	}

	for u := 0; u < alpha*numUsers; u++ {
		r := rand.Intn(30)
		if r == 0 {
			l.RequestsQueue <- l.GetLoginRequest(l.Names[u%numUsers])
		} else {
			if r%2 == 0 {
				bookID := l.Books[rand.Intn(len(l.Books))]
				l.RequestsQueue <- l.GetGetBookRequest(l.Names[u%numUsers], bookID)
			} else {
				bookID := l.Books[rand.Intn(len(l.Books))]
				l.RequestsQueue <- l.GetEditBookRequest(l.Names[u%numUsers], bookID)
			}

		}
	}
}

func (t *TestResult) addNewRequestType(typeName string) {
	rr := &RequestResult{
		StatusCodesHist: make(map[int]int),
		responseTimes:   make([]float64, 0),
	}
	t.Requests[typeName] = rr
}

func (t *TestResult) computeConcurrencyInfo(starts, ends []int64, firstRequestTime int64) {
	if len(starts) != len(ends) {
		panic(fmt.Errorf("starts and ends must have the same length"))
	}
	if len(starts) == 0 {
		panic(fmt.Errorf("starts is empty"))
	}
	if t.TestDuration < 1 {
		panic(fmt.Errorf("test duration is 0"))
	}
	t.ConcurrencyInfo = &ConcurrencyInfo{}

	var unitConvertor int64 = 100
	t.ConcurrencyInfo.HowManyUnitsIsASecond = float64(1000) / float64(unitConvertor)
	fmt.Println(t.TestDuration, "t.TestDuration")
	duration := int(t.TestDuration/unitConvertor) + int(1000/unitConvertor)
	units := make([]float64, duration)

	for i := 0; i < len(starts); i++ {
		start := int((starts[i] - firstRequestTime) / unitConvertor)
		end := int((ends[i] - firstRequestTime) / unitConvertor)
		// fmt.Println(start, end)
		for j := start; j < end; j++ {
			units[j]++
		}
	}

	v, e := stats.Max(units)
	if e != nil {
		panic(e)
	}
	t.ConcurrencyInfo.MaxConcurrentRequestsInOneUnit = v

	v, e = stats.Mean(units)
	if e != nil {
		panic(e)
	}
	t.ConcurrencyInfo.AverageConcurrentRequestsInOneUnit = v

	v, e = stats.Percentile(units, 95)
	if e != nil {
		panic(e)
	}
	t.ConcurrencyInfo.Percentile95 = v

	v, e = stats.Percentile(units, 99)
	if e != nil {
		panic(e)
	}
	t.ConcurrencyInfo.Percentile99 = v
}

// GetStats ...
func (l *LoadGenerator) GetStats() {
	var firstRequestTime = l.Requests[0].Start
	var lastRequestTime int64

	starts := make([]int64, 0)
	ends := make([]int64, 0)

	testResult := &TestResult{Unit: "ms"}
	testResult.Requests = make(map[string]*RequestResult)
	for _, r := range l.Requests {
		if _, ok := testResult.Requests[r.Type]; !ok {
			testResult.addNewRequestType(r.Type)
		}
		responseTime := float64(r.Finish - r.Start)
		starts = append(starts, r.Start)
		ends = append(ends, r.Finish)
		testResult.Requests[r.Type].responseTimes = append(testResult.Requests[r.Type].responseTimes, responseTime)

		if count, ok := testResult.Requests[r.Type].StatusCodesHist[r.StatusCode]; ok {
			testResult.Requests[r.Type].StatusCodesHist[r.StatusCode] = count + 1
		} else {
			testResult.Requests[r.Type].StatusCodesHist[r.StatusCode] = 1
		}

		testResult.Requests[r.Type].Count++

		if r.Start < firstRequestTime {
			firstRequestTime = r.Start
		}
		if r.Start > lastRequestTime {
			lastRequestTime = r.Start
		}
	}
	testResult.TestDuration = lastRequestTime - firstRequestTime
	for _, result := range testResult.Requests {
		testResult.TotalNumberOfRequests += result.Count

		v, err := stats.Mean(result.responseTimes)
		if err != nil {
			panic(err)
		}
		result.AverageResponseTime = v

		v, err = stats.Percentile(result.responseTimes, 95)
		if err != nil {
			panic(err)
		}
		result.Percentile95 = v

		v, err = stats.Percentile(result.responseTimes, 99)
		if err != nil {
			panic(err)
		}
		result.Percentile99 = v
	}
	testResult.computeConcurrencyInfo(starts, ends, firstRequestTime)
	l.result <- testResult
}

// Log ...
func Log(content string) {

}

// GetToken ...
func (l *LoadGenerator) GetToken(name string) string {
	l.TokensLock.RLock()
	defer l.TokensLock.RUnlock()
	return l.Tokens[name]
}

// WriteToken ...
func (l *LoadGenerator) WriteToken(name, token string) {
	l.TokensLock.Lock()
	defer l.TokensLock.Unlock()
	if len(l.Tokens[name]) > 4 {
		return
	}

	l.Tokens[name] = token
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

// Stop ...
func (l *LoadGenerator) Stop() {
	panic("Implement THIS!")
}
