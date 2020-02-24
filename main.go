package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	// lg := GetLoadGenerator("http://136.159.209.204:9080")
	// NumUsers := 100
	// Alpha := 50
	// NumWorkers := 30
	// lg.PrepareLoad(NumUsers, Alpha)
	// lg.GenerateLoad(NumWorkers)
	// fmt.Println(lg.Result)
	// time.Sleep(3 * time.Second)
	// fmt.Println("second round")
	// lg = GetLoadGenerator("http://136.159.209.204:9080")
	// lg.PrepareLoad(NumUsers, Alpha)
	// lg.GenerateLoad(NumWorkers)
	// fmt.Println(lg.Result)
	port := flag.String("port number", "7111", "port to listent to")
	c := GetController()
	fmt.Println("server started and listening to port", *port)
	http.ListenAndServe(":"+*port, http.HandlerFunc(c.Handler))
}
