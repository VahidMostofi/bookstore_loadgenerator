package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	port := flag.String("port number", "7111", "port to listent to")
	c := GetController()
	fmt.Println("server started and listening to port", *port)
	http.ListenAndServe(":"+*port, http.HandlerFunc(c.Handler))
}
