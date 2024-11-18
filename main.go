package main

import (
	"flag"
	"fmt"
	"net/http"
	"io"
	"os"
	"time"
)

type Request struct {
	TimeStart time.Time
	TimeEnd time.Time
	TimeDuration time.Duration
	DataVolume int
}

var requests []Request

func URLresponse(url string, n int)  {
	//var resp *http.Response
	for i := 1; i <= n; i++ {
		start := time.Now()
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
		end := time.Now()
		duration := end.Sub(start)
		req := Request{start, end, duration, len(body)}
		requests = append(requests, req)
		func () {
			defer resp.Body.Close()
		} ()
	}
	stdOut(requests)
}

func stdOut(sliceReq []Request) {
	timeFormat := "2006-01-02 15:04:05.999999999"
	fmt.Println("TimeStart 		     TimeEnd 			 TimeDuration  	DataVolume")
    fmt.Println("====================================================================================")
	for _, req := range sliceReq {
		startFormat := req.TimeStart.Format(timeFormat)
		endFormat := req.TimeEnd.Format(timeFormat)
		fmt.Printf("%-28v %-28v %14v %10v\n", startFormat, endFormat, req.TimeDuration, req.DataVolume)
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