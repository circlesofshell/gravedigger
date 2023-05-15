package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey

type graveUrlNode struct {
	graveUrlNode []graveUrl
}

type graveUrl struct {
	url        string
	httpStatus string
	//subDomain  string
}

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

// we can have a slice of graveUrl Type e.g. make([]graveUrl,0,len(FOOBAR))
// make is used to create dynamically-sized arrays
func checkStatus(urls []string) []graveUrl {

	//fill Struct in here you have the URLs and the resulting Status
	//empty structs in result because we make() with len of urls but some URLS are not reachable thus are not getting added to the slice of struct
	graveUrls := make([]graveUrl, len(urls))

	//a GET for ALL urls, skipping first entry as its always "original"
	for _, u := range urls[1:] {
		fmt.Println("Checking", u)
		resp, err := http.Get(u)

		if err != nil {
			continue
		}

		if resp == nil {
			continue
		} else {
			graveUrls = append(graveUrls, graveUrl{url: u, httpStatus: resp.Status})
		}

	}
	return graveUrls
}

func getSubdomain(u string) string {
	parse, err := url.Parse(u)

	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Split(parse.Hostname(), "."))

	return "string"
}

func main() {

	//httpClient := http.Client{Timeout: 2 * time.Second}
	//http://ie-drivermanagement-ieng-staging.ieng-staging.je-labs.com/

	getSubdomain("http://ie-drivermanagement-ieng-staging.ieng-staging.je-labs.com/")
	os.Exit(1)

	if len(os.Args) < 1 {
		fmt.Println("You need to supply a domain.")
		os.Exit(1)
	}
	urlToCheck := os.Args[1]

	archiveLink := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&collapse=urlkey", urlToCheck)

	urls := getUrls(archiveLink)

	statusMap := checkStatus(urls)

	for _, status := range statusMap {
		fmt.Println(status)

	}

}
