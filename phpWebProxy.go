package main

import (
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
)

type PhpWebProxy struct {
	WebProxy
	base64EncodeURL bool
}

func (pp *PhpWebProxy) init(proxyURL string, timeout int, header *Header, userAgent string) bool {
	if pp.Initialised == true {
		return true
	}
	if true {
		//php proxy doesn't work for now
		pp.SupportsPostRequest = false
		return false
	}
	pp.Transp.init("")
	pp.Initialised = true
	pp.Header = header

	pp.Header.init(proxyURL, userAgent)
	pp.Timeout = timeout
	var proxyPageContent string
	pp.LastLocation, _, proxyPageContent = GetRequest(proxyURL, pp.Transp.transport, pp.Timeout, pp.Header, 5)
	pp.SupportsPostRequest = false
	proxyAction := regexp.MustCompile(
		`form method="post" action="([^"]*)".{20,300}.{20,300}.{20,300}.{20,300}.{20,300}Use ROT13 encoding`).
		FindStringSubmatch(proxyPageContent)
	upars, _ := url.Parse(proxyURL)
	var paction string
	if proxyAction != nil {
		actPars, _ := upars.Parse(proxyAction[1])
		paction = actPars.String()
	}
	if strings.Contains(proxyPageContent, "base64 encodng") &&
		strings.Contains(proxyPageContent, "ROT13") &&
		strings.Contains(proxyPageContent, "scripting (i.e JavaScript)") &&
		strings.Contains(proxyPageContent, "Allow cookies") &&
		strings.Contains(proxyPageContent, "PHProxy") &&
		strings.Contains(paction, "http") {
		pp.SupportsPostRequest = true //might be changed further down
	}
	reg := regexp.MustCompile("</?form")
	allForms := reg.Split(proxyPageContent, -1)
	//fmt.Println(proxyPageContent)
	//loop through and test if this is the proxy's url form section
	for _, formContent := range allForms {
		//get action url, get all hidden inputs and values
		//if text input fields is more than one, skip this entry
		//find all inputs
		//fmt.Println(formContent)
		inRegexp := regexp.MustCompile(`<input[^>]+type="(?:text|password|email|phone|number)"[^>]+name="([^"]+)`)
		allTextInputs := inRegexp.FindAllStringSubmatch(formContent, -1)
		if allTextInputs == nil || len(allTextInputs) != 1 {
			//this form is most likely not for web address
			continue
		}
		//get action
		formContentLines := strings.Split(formContent, "\n")
		var action string
	forLoop:
		for _, line := range formContentLines {
			//<form method="post" action="/logs2/index.php">
			regi := regexp.MustCompile(`action="([^"]+)`)
			mts := regi.FindStringSubmatch(line)
			if mts != nil {
				action = mts[1]
				break forLoop
			}
		}
		//fmt.Println(action)
		//get all inputs fields names and values
		postData := allTextInputs[0][1] + "={url}&hl[accept_cookies]=on"
		postData += "&hl[show_images]=on&hl[show_referer]=on&hl[session_cookies]=on"

		pp.URL = proxyURL
		pp.PostData = postData
		pp.ActionURL = action
		if strings.Contains(pp.ActionURL, ".php") == false {
			pp.SupportsPostRequest = false
			return false
		}
		pp.resolveActionURL()

		pp.setRequisites(proxyURL, formContent)

	}
	return true
}

func (pp *PhpWebProxy) setRequisites(proxyURL string, formContent string) {
	//send a sample web request and check if urlencoding works
	plink := "https://www.google.com/"
	var params = postParameters{
		postLink: pp.ActionURL,
		postBody: strings.Replace(pp.PostData, "{url}", plink, -1),
		transp:   pp.Transp.transport,
		timeout:  pp.Timeout,
		header:   pp.Header}
	//not interested in output for now
	PostRequest(params, 5)
	pp.SupportsPostRequest = true
	pp.ProxyType = "Phproxy"
	if strings.Contains(pp.Header.LastLocation(), getURLHostname(plink)) == true {
		//url encodine allowed
		pp.EncodeMode = "urlencode"
		pp.base64EncodeURL = false
	} else if strings.Contains(pp.Header.LastLocation(),
		url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(plink)))) == true {
		//url encodine allowed
		pp.EncodeMode = "base64"
		pp.base64EncodeURL = true
	} else {
		//we can't use this proxy
		pp.SupportsPostRequest = false
	}

	//maybe also check for rot13 encoding
	//set up the struct instance accordingly
}

func (pp *PhpWebProxy) encryptURL(URLToEncrypt string) string {
	//remove http from it
	if pp.EncodeMode == "urlencode" || pp.base64EncodeURL != true {
		return url.QueryEscape(URLToEncrypt)
	}
	return base64.StdEncoding.EncodeToString([]byte(URLToEncrypt))
}

func (pp *PhpWebProxy) getRequest(requestURL string, proxy string, timeout int) string {

	var response string
	//send a post to first page form action just in case it is needed
	//first post request already made in setrequisites method
	reqURL := strings.Replace(pp.ActionURL, "{url}", pp.encryptURL(requestURL), -1)
	_, _, response = GetRequest(reqURL, pp.Transp.transport, timeout, pp.Header, 5)

	return response
}

func (pp *PhpWebProxy) postRequest(requestURL string, postBody string, proxy string, timeout int) string {
	//use cookie saved in object property
	//then send post request to proxy
	//include postmethod parameter in postdata
	var params = postParameters{
		postLink: strings.Replace(pp.ActionURL, "{url}", pp.encryptURL(requestURL), -1),
		postBody: strings.Replace(pp.PostData, "{url}", pp.encryptURL(requestURL), -1) + "&" + postBody,
		transp:   pp.Transp.transport,
		timeout:  timeout,
		header:   pp.Header,
	}
	//fmt.Println(params.postLink, params.postBody, pp.header.CookieJar.ToString(), "method=post")
	var response string
	//pp.proxyCookie(requestURL)
	pp.LastLocation, _, response = PostRequest(params, 5)
	//fmt.Println(params.postLink, pp.actionURL, pp.header.CookieJar.ToString() /* pp.header.Location.Archive, */, "method=post")
	return response
}
