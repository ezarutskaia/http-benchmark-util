package main

import (
	"bufio"
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
}

var requests []Request

func URLresponse(url string)  {
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
		req := Request{start, end, duration, len(body)}
		requests = append(requests, req)
}

func GOresponse(url string, n int)  {

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
		line := fmt.Sprintf("TimeStart: %s, TimeEnd: %s, TimeDuration: %s, DataVolume: %d",
			req.TimeStart, req.TimeEnd, req.TimeDuration, req.DataVolume)
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
	file := flag.Bool("file", false, "Safe log_file")
	parallel := flag.Bool("parallel", false, "Parallel execution of requests")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The 'url' parameter is required.")
		flag.Usage()
		os.Exit(1)
	}

	switch *parallel {
	case false:
		for i := 1; i <= *count; i++ {
			URLresponse(*url)
		}
	case true:
		wg.Add(*count)
		for i := 1; i <= *count; i++ {
			func () {
				defer wg.Done()
				go URLresponse(*url)
			} ()
		}
	}

	wg.Wait()
	stdOut(requests)

	if *file == true {
		SafeLog(*url, requests)
	}
}