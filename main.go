package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type waybackStatus struct {
	url        string
	httpStatus string
	//subDomains []string
}

func main() {

	var statusCode bool
	flag.BoolVar(&statusCode, "status", false, "display HTTP status code[!takes ages for a large set of URLs!]")
	var subDomains bool
	flag.BoolVar(&subDomains, "subdomains", false, "display found subdomains")
	var justUrls bool
	flag.BoolVar(&justUrls, "urls", false, "display only urls")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Printf("Usage:")
		fmt.Printf("\t" + "gravedigger -urls domain.tld to dig only for Urls\n")
		fmt.Printf("\t" + "gravedigger -subdomains domain.tld to print out subdomains\n")
		fmt.Printf("\t" + "gravedigger -status domain.tld to perform a HTTP status check on the URLS (WARNING:Can take a long ass time)\n")
		os.Exit(1)
	}

	urlToCheck := flag.Arg(0)

	archiveLink := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&collapse=urlkey", urlToCheck)

	urls := getUrls(archiveLink)

	if statusCode {
		checkStatus(urls[1:])
	}

	if justUrls {
		for _, url := range urls[1:] { //first item is always "original"
			fmt.Println(url)
		}
		os.Exit(0)
	}

	if subDomains {
		var sd []string
		for _, u := range urls[1:] {
			subs, err := getSubdomain(u)
			if err != nil {
				continue
			}
			sd = append(sd, subs[:1]...)

		}

		for _, sds := range removeDuplicates(sd) {
			fmt.Println(sds)
		}

		os.Exit(0)
	}

	//statusMap := checkStatus(urls)
	// TODO: move get subDomains out of this loop as only checkStatus can take very long
	// and its only if/else right now, so only checked for one flag

	/* 	flag.VisitAll(func (f *flag.Flag) {
		if f.Value.String()=="" {
			fmt.Println(f.Name, "not set!")
		}
	}) */

	/* for _, status := range statusMap {
		if status.url == "" {
			continue
		} else if statusCode {
			fmt.Println(status.url, status.httpStatus)
		}
	} */
}

// helper function to remove duplicates from subdomain slice
// TODO: works but ugly as shit. Why using a counter anyway?https://hackernoon.com/how-to-remove-duplicates-in-go-slices
func removeDuplicates(listOfSubDomains []string) []string {
	m := make(map[string]int)
	var singleSubs []string

	counter := 0

	for _, sub := range listOfSubDomains {
		m[sub] = counter
		if _, keyExists := m[sub]; keyExists {
			//log.Println("Subdomain already exists")
			counter++
		}
	}
	for k := range m {
		singleSubs = append(singleSubs, k)

	}
	//log.Println(m)
	return singleSubs
}

// TODO:some times parse fails e.g. INVALID URL ESCAPE "%"
// TODO: write test for checking www.just-eat.uk/% -> should not parse
// TODO: if hostname exclude from list
func getSubdomain(u string) ([]string, error) {
	parse, err := url.Parse(u)

	if err != nil {
		return nil, fmt.Errorf("An error occurred while parsing URl")

	}

	return strings.Split(parse.Hostname(), "."), nil
}

// TODO: checkStatus takes very long for a large set of data
/* func checkStatus(urls []string) []waybackUrl {
	start := time.Now()
	client := &http.Client{Timeout: timeout}

	//empty structs in result because we make() with len of urls but some URLS are not reachable thus are not getting added to the slice of struct
	waybackUrls := make([]waybackUrl, len(urls))

	//a GET for ALL urls, skipping first entry as its always "original"
	for _, u := range urls[1:] {

		subDomains, err := getSubdomain(u)
		if err != nil {
			continue
		}
		log.Println("Checking", u)
		resp, err := client.Get(u)
		// some Urls are not resolving
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp == nil {
			continue
		} else {
			waybackUrls = append(waybackUrls, waybackUrl{url: u, httpStatus: resp.Status, subDomains: subDomains})
		}

	}
	elapsed := time.Since(start)
	log.Printf("Function took %s", elapsed)
	return waybackUrls
} */

func checkStatus(urls []string) {
	//fmt.Println(urls)
	urlsChan := make(chan string, 100)
	statusChan := make(chan string)

	//log.Println("in checkStatus")

	var urlStatus []string

	for i := 0; i < cap(urlsChan); i++ {
		go checkStatusWorker(urlsChan, statusChan)
	}
	go func() {
		for _, url := range urls {
			urlsChan <- url
		}
	}()
	
	
	for i := 0; i < len(urls); i++ { //not running into a panic anymore because we are checking on the capacity of urls
		urlStat := <-statusChan
		if urlStat != "" {
			//correctly displays all the STATUS without panic?
			fmt.Println(urls[i],urlStat)
			urlStatus = append(urlStatus, urlStat)
		}

	}

	close(urlsChan)
	close(statusChan)

	//fmt.Println("checkstatus finished")
	/* for _, url := range urlStatus {
		fmt.Printf("%s what\n", url)
	} */

}

func checkStatusWorker(urls, status chan string) {
	const timeout = 1 * time.Second
	start := time.Now()
	client := &http.Client{Timeout: timeout}
	//log.Println("in Worker")

	//empty structs in result because we make() with len of urls but some URLS are not reachable thus are not getting added to the slice of struct
	//waybackUrls := make([]waybackStatus, len(urls))

	//a GET for ALL urls, skipping first entry as its always "original"
	for url := range urls {
		//log.Println("Checking", url)
		resp, err := client.Get(url)
		// some Urls are not resolving
		if err != nil {
			status <- ""
			continue
		}

		defer resp.Body.Close()
		if resp == nil {
			status <- ""
			continue
		} else {
			status <- resp.Status
			//waybackUrls = append(waybackUrls, waybackStatus{url: url, httpStatus: resp.Status})
		}

	}
	elapsed := time.Since(start)
	log.Printf("Function took %s", elapsed)

}

// TODO: Move getSubdomains into this function
func getUrls(archiveLink string) []string {
	var urlList []string

	//one request to get all the domains
	resp, err := http.Get(archiveLink)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	var wrapper [][]string //this is the format of the JSON we are GETting
	json.Unmarshal(body, &wrapper)

	for _, row := range wrapper {
		urlList = append(urlList, row[2]) //row[2] is the domain
	}

	return urlList
}
