package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

const (
	//DOP degree of Pararism
	DOP = 32
)

// A Result contains the title and URL of a search result.
type Result struct {
	domain, Status, serverTime, server, poweredBy string
}

func main() {

	domains := []string{"360.com", "adobe.com", "alibaba.com", "aliexpress.com", "amazon.co.jp", "amazon.co.uk", "amazon.com", "amazon.de",
		"amazon.in", "apple.com", "ask.com", "baidu.com", "bing.com", "blogger.com", "blogspot.com", "bongacams.com", "chase.com", "chinadaily.com.cn", "cnn.com",
		"coccoc.com", "craigslist.org", "diply.com", "dropbox.com", "ebay.com", "facebook.com", "fc2.com", "flipkart.com", "github.com", "gmw.cn", "go.com", "google.ca",
		"google.co.id", "google.co.in", "google.co.jp", "google.co.uk", "google.com", "google.com.au", "google.com.br", "google.com.hk", "google.com.mx", "google.com.tr",
		"google.de", "google.fr", "google.it", "google.pl", "google.ru", "hao123.com", "haosou.com", "imdb.com", "imgur.com", "instagram.com",
		"jd.com", "kat.cr", "linkedin.com", "live.com", "mail.ru", "microsoft.com", "microsoftonline.com", "msn.com", "Naver.com", "netflix.com", "nicovideo.jp",
		"nytimes.com", "office.com", "ok.ru", "onclickads.net", "outbrain.com", "PayPal.com", "pinterest.com", "pixnet.net", "pornhub.com", "qq.com", "rakuten.co.jp",
		"reddit.com", "sina.com.cn", "sohu.com", "soso.com", "stackoverflow.com", "t.co", "taobao.com", "tianya.cn", "tmall.com", "tumblr.com", "twitter.com", "vk.com",
		"weibo.com", "whatsapp.com", "wikipedia.org", "wordpress.com", "xhamster.com", "BAD NAME.com", "xvideos.com", "yahoo.co.jp", "yahoo.com", "yandex.ru", "youtube.com"}

	if err := fetchWebSites(domains); err != nil {
		log.Fatal(err)
	}

}

func fetchWebSites(domains []string) error {
	errChan := make(chan error, 1)
	sem := make(chan int, DOP) // N=DOP jobs at once
	// numTasks := len(domains)
	result := make(chan Result, DOP)

	var wg sync.WaitGroup

	go handleResult(result)

	for _, domain := range domains {
		wg.Add(1)
		go worker("http://"+domain, sem, &wg, errChan, result)
	}

	wg.Wait()
	close(errChan)
	return <-errChan
}

func handleResult(res chan Result) {
	for r := range res {
		fmt.Printf("%v\n", r)

	}
}

func worker(url string, sem chan int, wg *sync.WaitGroup, errChan chan error, res chan Result) {
	defer func() {
		<-sem
		wg.Done()
	}()

	sem <- 1

	//do the work
	resp, err := http.Get(url)
	//error handling
	if err != nil {
		select {
		case errChan <- err:
			// we're the first worker to fail
		default:
			//some other failure has already happened, discard this err by not sending it to the errChan
		}
		return
	}

	//process result
	defer resp.Body.Close()

	res <- Result{
		domain:     url,
		Status:     resp.Status,
		poweredBy:  resp.Header.Get("X-Powered-By"),
		server:     resp.Header.Get("Server"),
		serverTime: resp.Header.Get("Date"),
	}
}
