package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type postParameters struct {
	postLink string
	postBody string
	transp   *http.Transport
	timeout  int
	header   *Header
}

type stubMapping map[string]interface{}

var StubStorage = stubMapping{}

func writeToFile(filename string, data string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	_, err2 := f.WriteString(data + "\n")
	if err2 != nil {
		fmt.Println(err)
	}
}

func uniqueString(stringSlice *[]string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range *stringSlice {
		if value := keys[entry]; value != true {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

//ReadFile <
func ReadFile(filename string) []string {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		panic("error reading file: " + filename)
	}
	odat := strings.Replace(string(dat), "\r\n", "\n", -1)
	odat = strings.Replace(odat, "\r", "\n", -1)
	lines := strings.Split(odat, "\n")
	return uniqueString(&lines)
}

//CastByteToInteger <
func CastByteToInteger(value byte) int {
	retValue, _ := strconv.Atoi(string(value))
	return retValue
}

//CastToInteger <
func CastToInteger(value string) int {
	intValue, _ := strconv.Atoi(value)
	return intValue
}

//GetRequest <
func GetRequest(requestURL string, transp *http.Transport, timeout int, header *Header, maximumRedirects int) (string, string, string) {
	//timeout := time.Duration(30 * time.Second)
	ctx, cancel := context.WithCancel(context.Background()) //creates a timeout avenue
	defer cancel()
	//define client with proxy
	client := &http.Client{Transport: transp}

	//disable follow redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	req, er := http.NewRequestWithContext(
		ctx,
		"GET",
		requestURL, nil)
	//req.Close = true
	if er != nil {
		return requestURL, "", "error instantiating new request"
	}
	//req.Close = true
	contextCancelTimer := time.AfterFunc(time.Second*time.Duration(timeout), cancel)
	//is there authorization
	if len(header.Authorization) == 2 {
		req.SetBasicAuth(header.Authorization[0], header.Authorization[1])
	}
	header.setRequestHeaders(requestURL, req, "get")
	//req.Header["Accept-Encoding"] = []string{"deflate, gzip"}
	resp, err := client.Do(req)
	if err != nil {
		return requestURL, "", "error during get " + err.Error()
	}
	//still here? stop timer
	contextCancelTimer.Stop() //any recursive get request will have it's own context
	defer resp.Body.Close()

	header.saveHeaders(resp.Status, resp.Header)
	if len(resp.Header.Get("Location")) > 5 && maximumRedirects > 0 {
		//delegate to getrequest
		reqURLPas, _ := url.Parse(requestURL)
		redirLoc := resp.Header.Get("Location")
		nuURL, err := reqURLPas.Parse(redirLoc)
		if err == nil {
			return GetRequest(nuURL.String(), transp, timeout, header, maximumRedirects-1)
		}
		return redirLoc, "", "could not follow redirect"
	}
	//to stop ioutil from blocking forever

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte("error during read" + err.Error())
	}
	return requestURL, "", string(body)
	//return string(body)
}

//GetRequest <
func OptionRequest(requestURL string, transp *http.Transport, timeout int, header *Header, maximumRedirects int) (string, string, string) {
	//timeout := time.Duration(30 * time.Second)
	ctx, cancel := context.WithCancel(context.Background()) //creates a timeout avenue
	defer cancel()
	//define client with proxy
	client := &http.Client{Transport: transp}
	//disable follow redirects
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	req, er := http.NewRequestWithContext(
		ctx,
		"OPTIONS",
		requestURL, nil)
	//req.Close = true
	if er != nil {
		return requestURL, "", "error instantiating new request"
	}
	header.setRequestHeaders(requestURL, req, "option")
	//req.Close = true
	contextCancelTimer := time.AfterFunc(time.Second*time.Duration(timeout), cancel)
	//is there authorization
	if len(header.Authorization) == 2 {
		req.SetBasicAuth(header.Authorization[0], header.Authorization[1])
	}
	resp, err := client.Do(req)
	if err != nil {
		return requestURL, "", "error during get " + err.Error()
	}
	//still here? stop timer
	contextCancelTimer.Stop() //any recursive get request will have it's own context
	defer resp.Body.Close()

	header.saveHeaders(resp.Status, resp.Header)
	if len(resp.Header.Get("Location")) > 5 && maximumRedirects > 0 {
		//delegate to getrequest
		reqURLPas, _ := url.Parse(requestURL)
		redirLoc := resp.Header.Get("Location")
		nuURL, err := reqURLPas.Parse(redirLoc)
		if err == nil {
			return GetRequest(nuURL.String(), transp, timeout, header, maximumRedirects-1)
		}
		return redirLoc, "", "could not follow redirect"
	}
	//to stop ioutil from blocking forever

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte("error during read" + err.Error())
	}
	return requestURL, "", string(body)
	//return string(body)
}

//FileExists <
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//ReverseSlice <
func ReverseSlice(s interface{}) interface{} {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
	return s
}

func ReverseMapOfStrings(s map[string]string) map[string]string {
	var keys []string
	var values []string
	for k, v := range s {
		keys = append(keys, k)
		values = append(values, v)
	}
	retVal := map[string]string{}
	for x := len(keys) - 1; x >= 0; x-- {
		retVal[keys[x]] = values[x]
	}
	return retVal
}

func URLDecodeString(encoded string) string {
	dec, err := url.QueryUnescape(encoded)
	if err != nil {
		return encoded
	}
	return dec
}

//RegexpGetSubmatchFromString <
func RegexpGetSubmatchFromString(regex string, stringToSearch string) (string, error) {
	regie := regexp.MustCompile(regex)
	mts := regie.FindStringSubmatch(stringToSearch)
	if mts == nil {
		return "", fmt.Errorf("no match found for " + regex)
	}
	return mts[1], nil
}

//GetCookieFromHead <
func GetCookieFromHead(requestURL string, transp *http.Transport, timeout int) string {
	//timeout := time.Duration(30 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout)) //creates a timeout avenue
	defer cancel()
	//define client with proxy
	client := &http.Client{Transport: transp}

	req, er := http.NewRequest(
		"HEAD",
		requestURL, nil)
	if er != nil {
		return "error instantiating new request"
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.63 Safari/537.31")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "error during get"
	}
	/* cookieC := resp.Cookies()
	cookieString := ""
	//fmt.Println(resp.Header)
	for _, cooki := range cookieC {
		cookieString += cooki.Name + "=" + cooki.Value + ";"
	} */
	cookieString := ExtractCookieFromHeaders(resp.Header)
	defer resp.Body.Close()
	if err != nil {
		return "error during read"
	}
	return cookieString
}

func ExtractCookieFromHeaders(header http.Header) string {
	var cookie string
	cookies := header.Values("Set-Cookie")

	for _, cooki := range cookies {
		cookieList := strings.Split(cooki, ";")
		cookieParts := strings.Split(cookieList[0], "=") //first item is cookie rest is params
		if len(cookieParts) == 2 {
			cookie += fmt.Sprintf("%s=%s;", cookieParts[0], cookieParts[1])
		} else if len(cookieParts[0]) > 0 {
			cookie += fmt.Sprintf("%s=;", cookieParts[0])
		}
	}
	//fmt.Println(cookie)
	return cookie
}

func CombineCookies(cookieStringOne string, cookieStringTwo string) string {
	cookieTwoParts := strings.Split(cookieStringTwo, ";")
	for _, cookieItem := range cookieTwoParts {
		if strings.Contains(cookieStringOne, cookieItem) == false {
			//first remove any previous occurence
			ciParts := strings.Split(cookieItem, "=")
			reg := regexp.MustCompile(ciParts[0] + `=[^;]+`)
			cookieStringOne = reg.ReplaceAllString(cookieStringOne, "")
			cookieStringOne += ";" + cookieItem
		}
	}
	return cookieStringOne
}

//PostRequest <
func PostRequest(postParams postParameters, maximumRedirects int) (string, string, string) {
	postLink := postParams.postLink
	postBody := postParams.postBody
	transp := postParams.transp
	timeout := postParams.timeout
	header := postParams.header
	//fmt.Println(postLink)
	//post body is string like {"name":"john", "age":"twenty"}
	timeoutOne := timeout + 5
	ctx, cancel := context.WithCancel(context.Background()) //creates a timeout avenue
	defer cancel()
	//fmt.Println(postBody)
	//check if its json data
	var req *http.Request
	var er error
	if strings.Contains(postBody, "json::") {
		parts := strings.Split(postBody, "::")
		jsonData := []byte(parts[1])
		header.setCustomValue("Content-Length", fmt.Sprintf("%d", len(parts[1])))
		header.setCustomRequestContentType("application/json; charset=UTF-8")
		req, er = http.NewRequestWithContext(ctx, "POST", postLink, bytes.NewBuffer(jsonData))
		//fmt.Println(header.requestToString())
	} else {
		header.setCustomValue("Content-Length", fmt.Sprintf("%d", len(postBody)))
		header.setCustomRequestContentType("application/x-www-form-urlencoded; charset=UTF-8")
		req, er = http.NewRequestWithContext(ctx, "POST", postLink, strings.NewReader(postBody))
	}
	//req.Close = true
	if er != nil {
		return postLink, "", "error instantiating new post request"
	}
	client := &http.Client{Transport: transp}
	//below disables follow redirect
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }

	//fmt.Println("executed post request")
	header.setRequestHeaders(postLink, req, "post")
	contextCancelTimer := time.AfterFunc(time.Second*time.Duration(timeoutOne), cancel)
	if len(header.Authorization) == 2 {
		req.SetBasicAuth(header.Authorization[0], header.Authorization[1])
	}
	resp, err := client.Do(req)
	if err != nil {
		return postLink, "", "error during post " + err.Error()
	}
	contextCancelTimer.Stop()
	//fmt.Printf("%v\n", req)
	defer resp.Body.Close()
	header.saveHeaders(resp.Status, resp.Header)
	if len(resp.Header.Get("Location")) > 5 && maximumRedirects > 0 {
		//delegate to getrequest
		reqURLPas, _ := url.Parse(postLink)
		redirLoc := resp.Header.Get("Location")
		nuURL, err := reqURLPas.Parse(redirLoc)
		if err == nil {
			return GetRequest(nuURL.String(), transp, timeout, header, maximumRedirects-1)
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return postLink, "", "error during read" + err.Error()
	}
	//fmt.Println(string(body))
	loc, err := resp.Location()
	if err == nil {
		return loc.String(), "", string(body)
	}
	//return requestURL, cookieString, body
	return postLink, "", string(body)
}

//StringInSlice <
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//ListElementsInString <
func ListElementsInString(listOfItems []string, stringToSearch string) bool {
	for _, item := range listOfItems {
		if strings.Contains(stringToSearch, item) == false {
			return false
		}
	}
	return true
}
