package main

import (
	"flag"
	"fmt"
	"net/http"
	"io"
)

func URLresponse(url string, n int)  {
	//var resp *http.Response
	for i := 1; i <= n; i++ {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s\n", body)
		func () {
			defer resp.Body.Close()
		} ()
	}
}

func main() {
	url := flag.String("url", "http://localhost:8080/", "url adress")
	count := flag.Int("count", 1, "Number of requests")
	flag.Parse()
	URLresponse(*url, *count)
}