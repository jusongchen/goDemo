package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// A Result contains the title and URL of a search result.
type taskResult struct {
	Status, serverTime, server, poweredBy string
	err                                   error
}

//Result exported
type Result struct {
	url        string
	taskResult *taskResult
	elapsed    time.Duration
	err        error
}

const (
	topN = 10
)

func main() {

	domains := []string{"salesforce.com", "360.cn", "360.com", "adobe.com", "alibaba.com", "aliexpress.com", "amazon.co.jp", "amazon.co.uk", "amazon.com", "amazon.de",
		"amazon.in", "apple.com", "ask.com", "baidu.com", "bing.com", "blogger.com", "blogspot.com", "bongacams.com", "chase.com", "chinadaily.com.cn", "cnn.com",
		"coccoc.com", "craigslist.org", "diply.com", "dropbox.com", "ebay.com", "facebook.com", "fc2.com", "flipkart.com", "github.com", "gmw.cn", "go.com", "google.ca",
		"google.co.id", "google.co.in", "google.co.jp", "google.co.uk", "google.com", "google.com.au", "google.com.br", "google.com.hk", "google.com.mx", "google.com.tr",
		"google.de", "google.fr", "google.it", "google.pl", "google.ru", "googleusercontent.com", "hao123.com", "haosou.com", "imdb.com", "imgur.com", "instagram.com",
		"jd.com", "kat.cr", "linkedin.com", "live.com", "mail.ru", "microsoft.com", "microsoftonline.com", "msn.com", "Naver.com", "netflix.com", "nicovideo.jp",
		"nytimes.com", "office.com", "ok.ru", "onclickads.net", "outbrain.com", "PayPal.com", "pinterest.com", "pixnet.net", "pornhub.com", "qq.com", "rakuten.co.jp",
		"reddit.com", "sina.com.cn", "sohu.com", "soso.com", "stackoverflow.com", "t.co", "taobao.com", "tianya.cn", "tmall.com", "tumblr.com", "twitter.com", "vk.com",
		"weibo.com", "whatsapp.com", "wikipedia.org", "wordpress.com", "xhamster.com", "BAD NAME.com", "xvideos.com", "yahoo.co.jp", "yahoo.com", "yandex.ru", "youtube.com"}

	// timeout := time.Duration(3000) * time.Millisecond
	timeout := 300 * time.Millisecond
	numURL := len(domains)

	results := make(chan Result, numURL) //buffered

	wgTopN, wgAll := sync.WaitGroup{}, sync.WaitGroup{}
	wgTopN.Add(topN)
	wgAll.Add(numURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, domain := range domains {
		go func(domain string) {
			start := time.Now()
			url := "http://" + domain

			taskResult, err := execTask(ctx, timeout, url)
			res := Result{elapsed: time.Since(start),
				taskResult: &taskResult,
				url:        url,
				err:        err,
			}
			results <- res
			if err == nil {
				wgTopN.Done()
			}
			wgAll.Done()
		}(domain)
	}
	//when hits timeout or get enough results,cancel all requests
	go func() {
		done := make(chan struct{})
		go func() {
			wgTopN.Wait()
			done <- struct{}{}
		}()

		select { //wait until one of the following events
		case <-time.After(timeout):
		case <-done:
		}
		cancel()
	}()

	//when all httpGet returns, close both channels
	go func() {
		wgAll.Wait()
		close(results)
	}()
	printResult(results, topN, numURL)
}

//printResult prints winners, then losers
func printResult(results <-chan Result, topN, numURL int) {
	losers := []Result{}

	i := 0
	for res := range results {
		if res.err == nil && i < topN {
			fmt.Printf("%v\t%-50s\t%v\n", res.elapsed, res.url, *res.taskResult)
			i++
		} else {
			losers = append(losers, res)
		}
	}

	fmt.Printf("\nLosers:\n")
	for _, res := range losers {
		fmt.Fprintf(os.Stderr, "%v\t%-50s\t%v\n", res.elapsed, res.url, res.err)
	}
}

// Search sends query to Google search and returns the results.
func execTask(parentCtx context.Context, timeout time.Duration, url string) (taskResult, error) {
	// Prepare the Google Search API request.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	c := make(chan taskResult, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			c <- taskResult{err: err}
			return
		}
		defer resp.Body.Close()

		c <- taskResult{
			Status:     resp.Status,
			poweredBy:  resp.Header.Get("X-Powered-By"),
			server:     resp.Header.Get("Server"),
			serverTime: resp.Header.Get("Date"),
		}
		close(c) //this will be executed even execTask returns first
	}()
	select {
	case <-ctx.Done():
		// log.Printf("httpDO:%s Get Done signal", req.URL)
		// <-c // Wait for f to return.
		return taskResult{}, ctx.Err()
	case res := <-c:
		return res, res.err
	}
}
