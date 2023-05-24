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

const timeout = 1 * time.Second

type waybackUrl struct {
	url        string
	httpStatus string
	subDomains []string
}

// TODO: Move getSubdomains into this function
func getUrls(archiveLink string) []string {
	var urlList []string
	fmt.Println(archiveLink)
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
		//fmt.Println(row[1])
		urlList = append(urlList, row[2]) //row[2] is the domain
	}

	return urlList
}

// TODO: checkStatus takes very long for a large set of data
// maybe only check the actuall domain?
// we can have a slice of waybackUrl Type e.g. make([]waybackUrl,0,len(FOOBAR))
// make is used to create dynamically-sized arrays
func checkStatus(urls []string) []waybackUrl {
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
}

// TODO:some times parse fails e.g. INVALID URL ESCAPE "%"
// TODO: write test for checking www.just-eat.uk/% -> should not parse
// Return error in this function and test for it in CHECK status -> continue
func getSubdomain(u string) ([]string, error) {
	parse, err := url.Parse(u)

	if err != nil {
		return nil, fmt.Errorf("An error occurred while parsing URl")

	}

	return strings.Split(parse.Hostname(), "."), nil
}

func main() {

	var statusCode bool
	flag.BoolVar(&statusCode, "status", false, "display HTTP status code")
	var subDomains bool
	flag.BoolVar(&subDomains, "subdomain", false, "display found subdomains")
	var justUrls bool
	flag.BoolVar(&justUrls, "urls", false, "display only urls")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("You need to supply a domain.")
		os.Exit(1)
	}

	urlToCheck := flag.Arg(0)

	archiveLink := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&collapse=urlkey", urlToCheck)

	urls := getUrls(archiveLink)

	if justUrls {
		for _, url := range urls[1:] { //first item is always "original"
			fmt.Println(url)
		}
		os.Exit(0)
	}

	statusMap := checkStatus(urls)
	// TODO: move get subDomains out of this loop as only checkStatus can take very long
	for _, status := range statusMap {
		if status.url == "" {
			continue
		} else if statusCode {
			fmt.Println(status.url, status.httpStatus)
		} else if subDomains {
			fmt.Println(status.subDomains)
		}
	}

}
