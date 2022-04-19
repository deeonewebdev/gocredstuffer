package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type WebProxyInterface interface {
	init(proxyURL string, timeout int, header *Header, userAgent string) bool
	supportsPostRequest() bool
	encryptURL(URLToEncrypt string) string
	getRequest(requestURL string, proxy string, timeout int) string
	postRequest(requestURL string, postBody string, proxy string, timeout int) string
	proxyType() string
	getHeader() *Header
	isInitialised() bool
}

type WebProxy struct {
	URL                 string
	ActionURL           string
	HotRequestURL       string
	PostData            string
	Header              *Header
	Transp              Transport
	LastLocation        string
	ProxyType           string
	EncodeMode          string
	Timeout             int
	SupportsPostRequest bool
	Initialised         bool
	_WebProxy           WebProxyInterface
}

func (wp *WebProxy) resolveActionURL() {
	//urlByteSlice := []byte(wp.actionURL)
	urlParts, err := url.Parse(wp.URL)
	if err != nil {
		return
	}
	//fmt.Println(wp.URL, wp.actionURL)
	absURL, err := urlParts.Parse(wp.ActionURL)
	if err == nil {
		wp.ActionURL = absURL.String()
	}
}

func (wp *WebProxy) proxyType() string {
	return wp.ProxyType
}

func (wp *WebProxy) getHeader() *Header {
	return wp.Header
}

func (wp *WebProxy) isInitialised() bool {
	return wp.Initialised
}
func (wp *WebProxy) supportsPostRequest() bool {
	return wp.SupportsPostRequest
}

func getAllInputNames(formContent string) []string {
	allInpReg := regexp.MustCompile("<input[^>]+")
	allInputs := allInpReg.FindAllString(formContent, -1) //all forms[1] is first index
	namesOfInputs := []string{}
	for _, inputLine := range allInputs {
		//get input name
		regnm := regexp.MustCompile(`name="([^"]+)`)
		mts := regnm.FindStringSubmatch(inputLine)
		if mts != nil {
			namesOfInputs = append(namesOfInputs, mts[1])
		}
	}
	return namesOfInputs
}

func getListDifference(newList []string, oldList []string) []string {
	retList := []string{}
	for _, item := range newList {
		if StringInSlice(item, oldList) == false {
			retList = append(retList, item)
		}
	}
	return retList

}

/* func (wp *WebProxy) setPostMethodParameter(inputs []string, formString string) {
	proxySpecificPostData := ""
	allInpReg := regexp.MustCompile("<input[^>]+")
	allInputs := allInpReg.FindAllString(formString, -1)
	for _, inputLine := range allInputs {
		for _, inputItem := range inputs {
			if strings.Contains(strings.ToLower(inputLine), strings.ToLower(inputItem)) {
				//get value
				rg := regexp.MustCompile(`value="?([^"\s]+)`)
				mts := rg.FindStringSubmatch(inputLine)
				if mts != nil {
					proxySpecificPostData += inputItem + "=" + mts[1] + "&"
				} else {
					proxySpecificPostData += inputItem + "=&"
				}
			}
		}
	}
	wp.postMethodParameter = proxySpecificPostData

} */

func addTrailingSlash(pURL string) string {
	uParts := strings.Split(pURL, "/")
	li := len(uParts) - 1
	if len(uParts) > 0 && strings.Contains(uParts[li], ".") == false {
		return pURL + "/"
	}
	return pURL
}

func getURLHostname(requestURL string) string {
	upars, _ := url.Parse(requestURL)
	hostNreg := regexp.MustCompile(`([^.]+\.[^.]+)$`)
	mts := hostNreg.FindStringSubmatch(upars.Host)
	if mts == nil {
		return ""
	}
	hostname := mts[1]
	return hostname
}
func (wp *WebProxy) proxyCookie(requestURL string) {
	upars, _ := url.Parse(requestURL)
	hostNreg := regexp.MustCompile(`([^.]+\.[^.]+)$`)
	mts := hostNreg.FindStringSubmatch(upars.Host)
	if mts == nil {
		return
	}
	hostname := mts[1]
	for k, v := range wp.Header.CookieJar.cookieMap {
		if strings.Contains(k, "]") == false &&
			k != "s" &&
			wp.ProxyType == "glype" {
			//key needs to pass through proxy
			delete(wp.Header.CookieJar.cookieMap, k)

			wp.Header.CookieJar.cookieMap[fmt.Sprintf("c[%s][/][%s]", hostname, k)] = v
		}
	}
}

func (wp *WebProxy) Reset() {
	//resets properties to default to allow
	//object to be reused for another request
}
