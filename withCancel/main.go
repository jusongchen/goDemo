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
type httpResult struct {
	Status, serverTime, server, poweredBy string
}

type Result struct {
	url        string
	httpResult *httpResult
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
		"weibo.com", "whatsapp.com", "wikipedia.org", "wordpress.com", "xhamster.com", "xinhuanet.com", "xvideos.com", "yahoo.co.jp", "yahoo.com", "yandex.ru", "youtube.com"}

	// timeout := time.Duration(3000) * time.Millisecond
	timeout := 500 * time.Millisecond
	numURL := len(domains)

	results := make(chan Result, topN)
	errResults := make(chan Result, numURL)

	wgTopN := sync.WaitGroup{}
	wgTopN.Add(topN)

	wgAll := sync.WaitGroup{}
	wgAll.Add(numURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, domain := range domains {
		go func(domain string) {
			start := time.Now()
			url := "http://" + domain

			httpRes, err := httpGet(ctx, timeout, url)
			res := Result{elapsed: time.Since(start),
				httpResult: httpRes,
				url:        url,
				err:        err,
			}
			if err != nil {
				//error handling
				errResults <- res
			} else {
				results <- res
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
		close(errResults)
	}()

	i := 0
	fmt.Printf("Winners:\n")
	for res := range results {
		if i == topN {
			fmt.Printf("\nLosers:\n")
		}
		fmt.Printf("%v\t\t%s\t\t%v\n", res.elapsed, res.url, res.httpResult)
		i++
	}

	fmt.Printf("\nDisqualified:\n")
	for res := range errResults {
		fmt.Fprintf(os.Stderr, "%v\t%s\t%v\n", res.elapsed, res.url, res.err)
	}
}

// Search sends query to Google search and returns the results.
func httpGet(parentCtx context.Context, timeout time.Duration, url string) (*httpResult, error) {
	// Prepare the Google Search API request.
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Issue the HTTP request and handle the response. The httpDo function
	// cancels the request if ctx.Done is closed.
	var result *httpResult

	handleFunc := func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		result = &httpResult{
			Status:     resp.Status,
			poweredBy:  resp.Header.Get("X-Powered-By"),
			server:     resp.Header.Get("Server"),
			serverTime: resp.Header.Get("Date"),
		}
		return nil
	}
	err = httpDo(ctx, req, handleFunc)
	// httpDo waits for the closure we provided to return, so it's safe to
	// read results here.
	return result, err
}

// httpDo issues the HTTP request and calls f with the response. If ctx.Done is
// closed while the request or f is running, httpDo cancels the request, waits
// for f to exit, and returns ctx.Err. Otherwise, httpDo returns f's error.
func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() {
		c <- f(client.Do(req))
	}()
	select {
	case <-req.Cancel:
		//the server cancelled request
		return fmt.Errorf("httpDo:request cancelled, %s", req.URL)
	case <-ctx.Done():
		// log.Printf("httpDO:%s Get Done signal", req.URL)
		tr.CancelRequest(req)
		// <-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
