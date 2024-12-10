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

const serverFallDown = "Server returnd status 500"
var requests []Request
var wg sync.WaitGroup
var mu sync.Mutex

func URLresponse(url string, pack int) error {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Request %d done with error: %v", pack, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf(serverFallDown)
	}

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
	return nil
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
	url := flag.String("url", "", "Url adress")
	count := flag.Int("count", 1, "Number of requests")
	bunch := flag.Int("bunch", 1, "Number of package splits, Parallel execution only")
	file := flag.Bool("file", false, "Safe log_file")
	parallel := flag.Bool("parallel", false, "Parallel execution of requests")
	timeout := flag.Int("timeout", 0, "Number of seconds of waiting for response")
	kill := flag.Bool("kill", false, "kill server")

	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The 'url' parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	duration := time.Duration(*timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	if *kill == true {
		*parallel = true
	//	*bunch = 10
	}

	switch *parallel {
	case false:
		for i := 1; i <= *count; i++ {
		    if err := URLresponse(*url, 1); err != nil {
	        	fmt.Println("Error:", err)
	    	}
		}
	case true:
		if *kill == false {
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
							if err := URLresponse(*url, 1); err != nil {
								fmt.Println("Error:", err)
							}
						} ()
					} else {
						go func (start int) {
							defer wg.Done()
							URLresponseTimeout(ctx, *url, start)
						} (start)
					}
				}
				wg.Wait()
			}
		} else {
			stop := make(chan struct{})
			i := 0
			outerLoop: for {
				select {
				case <-stop:
					fmt.Println("Stopping testing: server returned 500")
					break outerLoop
				default:
					i++
					wg.Add(1)
					go func(i int, stop chan struct{}) {
						fmt.Println("start goroutine", i)
						defer wg.Done()
						err := URLresponse(*url, i)
						if err != nil && err.Error() == serverFallDown {
							fmt.Printf("Error 500 received on request %d. Stopping all tests. Error %v\n", i, err)
							close(stop)
						}
					}(i, stop)

					wg.Wait()
				}
			}
		}
	}

	stdOut(requests)

	if *file == true {
		SafeLog(*url, requests)
	}
}