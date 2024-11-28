package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Request struct {
	TimeStart time.Time
	TimeEnd time.Time
	TimeDuration time.Duration
	DataVolume int
	Pack int
}

var requests []Request
var mu sync.Mutex

func URLresponse(url string, pack int)  {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	end := time.Now()
	duration := end.Sub(start)
	result := Request{start, end, duration, len(body), pack}
	mu.Lock()
	requests = append(requests, result)
	mu.Unlock()
}

func URLresponseTimeout(ctx context.Context,url string, pack int)  {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("Package: %d - Context cancelled: %v\n", pack, ctx.Err())
		return
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		end := time.Now()
		duration := end.Sub(start)
		result := Request{start, end, duration, len(body), pack}
		mu.Lock()
		requests = append(requests, result)
		mu.Unlock()
	}
}

func stdOut(sliceReq []Request) {
	timeFormat := "2006-01-02 15:04:05.999999999"
	fmt.Println("TimeStart 		     TimeEnd 			 TimeDuration  	DataVolume	Pack")
    fmt.Println("===============================================================================================")
	for _, req := range sliceReq {
		startFormat := req.TimeStart.Format(timeFormat)
		endFormat := req.TimeEnd.Format(timeFormat)
		fmt.Printf("%-28v %-28v %14v %10v %10v\n", startFormat, endFormat, req.TimeDuration, req.DataVolume, req.Pack)
	}
}

func SafeLog(url string, sliceReq []Request) {
	currentTime := time.Now()
	timeFormat := currentTime.Format("2006-01-02_15-04-05")
	dir := "./logs"
	safeURL := strings.ReplaceAll(url, "/", "_")
	safeURL = strings.ReplaceAll(safeURL, ":", "-")
	filename := safeURL + "_" + timeFormat + ".log"
	filePath := filepath.Join(dir, filename)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error create dir:", err)
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error create file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, req := range sliceReq {
		line := fmt.Sprintf("TimeStart: %s, TimeEnd: %s, TimeDuration: %s, DataVolume: %d, Pack: %d",
			req.TimeStart, req.TimeEnd, req.TimeDuration, req.DataVolume, req.Pack)
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flush:", err)
		return
	}
}

func main() {
	var wg sync.WaitGroup

	url := flag.String("url", "", "Url adress")
	count := flag.Int("count", 1, "Number of requests")
	bunch := flag.Int("bunch", 1, "Number of package splits, Parallel execution only")
	file := flag.Bool("file", false, "Safe log_file")
	parallel := flag.Bool("parallel", false, "Parallel execution of requests")
	timeout := flag.Int("timeout", 0, "Number of seconds of waiting for response")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The 'url' parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	duration := time.Duration(*timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	switch *parallel {
	case false:
		for i := 1; i <= *count; i++ {
			URLresponse(*url, 1)
		}
	case true:
		for start := 0; start < *count; start += *bunch {
			end := start + *bunch
			if end > *count {
				end = *count
			}

			for i := start; i < end; i++ {
				wg.Add(1)
				if *timeout == 0 {
					go func () {
						defer wg.Done()
						URLresponse(*url, start)
					} ()
				} else {
					go func () {
						defer wg.Done()
						URLresponseTimeout(ctx, *url, start)
					} ()
				}
			}
			wg.Wait()
		}
	}

	stdOut(requests)

	if *file == true {
		SafeLog(*url, requests)
	}
}