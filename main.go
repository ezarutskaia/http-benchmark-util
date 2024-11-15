package main

import (
	"flag"
	"fmt"
	"net/http"
	"io"
	"os"
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
	url := flag.String("url", "", "url adress")
	count := flag.Int("count", 1, "Number of requests")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The 'url' parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	URLresponse(*url, *count)
}