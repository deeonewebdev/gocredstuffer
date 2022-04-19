package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type GlypeWebProxy struct {
	WebProxy
	UseUnderscore bool
	PathMode      bool
	PostedTo      bool
	UniqueSalt    string
}

func (gp *GlypeWebProxy) setRequisites(requestURL string, contentFromProxy string) {

	gp.SupportsPostRequest = false
	//get variables from ginf object
	reg := make(map[string]*regexp.Regexp)
	value := make(map[string]string)
	reg["proxyHost"] = regexp.MustCompile(`url:'([^']*)`)
	reg["script"] = regexp.MustCompile(`script:'([^']*)`)
	reg["salt"] = regexp.MustCompile(`{u:'([^']*)`)
	reg["pathMode"] = regexp.MustCompile(`,p:'([^']*)'\}`)
	reg["encodeURL"] = regexp.MustCompile(`,e:'([^']*)`)

	var line string
	contentParts := strings.Split(contentFromProxy, "\n")
	//get glype config line
	ginfFound := false
forLoop:
	//gets and set line to one containing ginf=
	for _, line = range contentParts {
		if strings.Contains(line, "ginf={") {
			ginfFound = true
			break forLoop
		}
	}

	//fmt.Println(value, contentFromProxy)
	if ginfFound == false {
		return
	}
	if len(line) == 0 {
		gp.SupportsPostRequest = false
		return
	}

	//match variables
	for k, v := range reg {
		mts := v.FindStringSubmatch(line)
		value[k] = "" //initital value
		if mts != nil {
			value[k] = mts[1]
		}
	}
	gp.ProxyType = "glype"
	gp.PathMode = false
	gp.SupportsPostRequest = strings.Contains(contentFromProxy, value["proxyHost"])
	gp.UniqueSalt = value["salt"]
	gp.URL = addTrailingSlash(value["proxyHost"])
	gp.ActionURL = fmt.Sprintf("%s/%s?u={url}&b=4", gp.URL, value["script"])
	//indicates this proxy behaves properly

	//=https%3A%2F%2Fnaira4dollar.com%2F
	regURLE := regexp.MustCompile(`=https?%3A%2F%2F[^/%\s]+%2F[^/%\s]+`)
	matches := regURLE.Split(contentFromProxy, 100)
	//fmt.Println(gp.URL, gp.uniqueSalt, ">>>>>")
	if len(value["pathMode"]) > 0 {
		gp.PathMode = true
		gp.ActionURL = fmt.Sprintf("%s/%s/{url}/b4/", value["proxyHost"], value["script"])
	}
	gp.EncodeMode = "arcfour"
	if len(gp.UniqueSalt) < 10 || len(matches) > 50 {
		gp.EncodeMode = "urlencode"
	}
	//fmt.Println(gp.EncodeMode, value, len(matches))

}

func arcFour(mode string, uniqueSalt string, urlWithoutHTTP string) string {
	if mode == "decrypt" {
		URLbytes, _ := base64.StdEncoding.DecodeString(urlWithoutHTTP)

		urlWithoutHTTP = string(URLbytes)
	}
	growingURLString := ""
	zeroTo255 := make([]int, 256)
	twoFiveSix := 256
	saltLength := len(uniqueSalt)
	urlLength := len(urlWithoutHTTP)
	uniqueSaltBytes := []byte(uniqueSalt)
	for x := 0; x < len(zeroTo255); x++ {
		zeroTo255[x] = x
	}

	zeroJ := 0
	for i := 0; i < len(zeroTo255); i++ {
		zeroJ = (zeroJ + zeroTo255[i] + int(uniqueSaltBytes[i%saltLength])) % twoFiveSix
		x := zeroTo255[i]
		zeroTo255[i] = zeroTo255[zeroJ]
		zeroTo255[zeroJ] = x
	}

	//fmt.Println(zeroTo255)
	zeroJ2 := 0
	zeroI := 0
	growBytes := []byte{}
	for y := 0; y < urlLength; y++ {
		zeroI = (zeroI + 1) % twoFiveSix
		zeroJ2 = (zeroJ2 + zeroTo255[zeroI]) % twoFiveSix
		x := zeroTo255[zeroI]
		zeroTo255[zeroI] = zeroTo255[zeroJ2]
		zeroTo255[zeroJ2] = x
		secChar := zeroTo255[(zeroTo255[zeroI]+zeroTo255[zeroJ2])%twoFiveSix]
		ch := byte(urlWithoutHTTP[y]) ^ byte(secChar)
		growBytes = append(growBytes, ch)
	}

	growingURLString = string(growBytes)
	if mode == "encrypt" {
		growingURLString = base64.StdEncoding.EncodeToString([]byte(growingURLString))
	}
	return growingURLString

}

func (gp *GlypeWebProxy) agreeSSL(response string) (string, string, string) {
	//find the action link
	reg := regexp.MustCompile(`form action="([^"]+)" method="get"`)
	mts := reg.FindStringSubmatch(response)
	if mts != nil {
		fullURL, err := url.Parse(gp.URL)
		if err == nil {
			fullURL, err = fullURL.Parse(mts[1])
			if err == nil {
				actURL := fullURL.String() + "?action=sslagree"
				lloc, cook, res := GetRequest(actURL, gp.Transp.transport, gp.Timeout, gp.Header, 5)
				return lloc, cook, res
			}
		}
	}
	return "", "", response
}

func (gp *GlypeWebProxy) encryptURL(URLToEncrypt string) string {
	//remove http from it
	if gp.EncodeMode == "urlencode" {
		return url.QueryEscape(URLToEncrypt)
	}
	URLRunes := []rune(URLToEncrypt)
	urlWithoutHTTP := string(URLRunes[4:])
	//fmt.Println(urlWithoutHTTP, "::::::", gp.uniqueSalt, gp.URL)
	arcURL := arcFour("encrypt", gp.UniqueSalt, urlWithoutHTTP)
	URLEncodedArcURL := url.QueryEscape(arcURL)
	//replace % with _
	if gp.UseUnderscore == true {
		URLEncodedArcURL = strings.Replace(URLEncodedArcURL, "%", "_", -1)
	}
	//put / in every 8 character if it's path encoded
	if gp.PathMode == true {
		reg := regexp.MustCompile("(.{8})")
		eightChunk := reg.Split(URLEncodedArcURL, -1)
		URLEncodedArcURL = strings.Join(eightChunk, "/")
	}
	return URLEncodedArcURL
}

func (gp *GlypeWebProxy) getRequest(requestURL string, proxy string, timeout int) string {

	//use cookie saved in object property
	//then send post request to proxy to fetch the page
	var response string
	//fmt.Println(requestURL, "called")
	//should only do this once
	if gp.PostedTo != true {
		gp.PostedTo = true
		//send a post to first page form action just in case it is needed
		_, _, contentFromProxy := GetRequest(gp.URL, gp.Transp.transport, timeout, gp.Header, 5)
		//fmt.Println(gp.header.CookieJar.ToString(), "==========", gp.URL)
		actionReg := regexp.MustCompile(`action="([^"]+)"[^\n\r]+return updateLocation`)
		mts := actionReg.FindStringSubmatch(contentFromProxy)
		if mts != nil {
			upars, _ := url.Parse(gp.URL)
			actPars, _ := upars.Parse(mts[1])
			actURL := actPars.String()
			//send a post request to the action url
			posp := postParameters{
				postLink: actURL,
				postBody: fmt.Sprintf("u=%s&allowCookies=on", url.QueryEscape(requestURL)),
				transp:   gp.Transp.transport,
				timeout:  gp.Timeout,
				header:   gp.Header}
			//not interested in output
			_, _, resp := PostRequest(posp, 5)
			if strings.Contains(strings.ToLower(resp), `value="sslagree"`) {
				//agree and send of form and overwrite rawContent
				_, _, resp = gp.agreeSSL(resp)
			}
			//check if url is encoded
			gp.setRequisites(requestURL, resp)
			if strings.Contains(gp.Header.LastLocation(), getURLHostname(requestURL)) == false {
				//set encrypt mode to true
				//gp.EncodeMode = "arcfour"
				if strings.Contains(gp.Header.LastLocation(), "&b=") == false {
					//set glypparamse pathmode
					gp.PathMode = true
				}
			}
			gp.UseUnderscore = false
			if strings.Contains(gp.Header.LastLocation(), "%") == false {
				gp.UseUnderscore = true
			}
		}
	}
	reqURL := strings.Replace(gp.ActionURL, "{url}", gp.encryptURL(requestURL), -1)
	//gp.proxyCookie(requestURL)
	gp.LastLocation, _, response = GetRequest(reqURL, gp.Transp.transport, timeout, gp.Header, 5)

	//check if its glype proxy and its asking for sslagree
	if strings.Contains(strings.ToLower(response), `value="sslagree"`) {
		//agree and send of form and overwrite rawContent
		gp.LastLocation, _, response = gp.agreeSSL(response)
	}
	return response
}

func (gp *GlypeWebProxy) postRequest(requestURL string, postBody string, proxy string, timeout int) string {
	//use cookie saved in object property
	//then send post request to proxy
	//include postmethod parameter in postdata
	if gp.SupportsPostRequest == false {
		fmt.Println("-web proxy is bad " + gp.URL)
		return "bad web proxy"
	}
	var params = postParameters{
		postLink: strings.Replace(gp.ActionURL, "{url}", gp.encryptURL(requestURL), -1),
		postBody:/* strings.Replace(gp.PostData, "{url}", gp.encryptURL(requestURL), -1) + "&" +  */ postBody,
		transp:  gp.Transp.transport,
		timeout: timeout,
		header:  gp.Header,
	}
	//fmt.Println(params.postLink, params.postBody, gp.header.CookieJar.ToString(), "method=post")
	_, _, response := PostRequest(params, 5)
	pur, _ := url.Parse(gp.URL)
	if strings.Contains(response, pur.Host) == false {
		gp.SupportsPostRequest = false
		fmt.Println("web proxy is bad "+gp.URL, requestURL)
		return "bad web proxy "
	}
	if strings.Contains(gp.URL, "kovals.eu") {
		fmt.Println(response)
	}
	//check if its glype proxy and its asking for sslagree
	if strings.Contains(strings.ToLower(response), `value="sslagree"`) {
		//agree and send of form and overwrite rawContent
		gp.LastLocation, _, response = gp.agreeSSL(response)
	}
	//fmt.Println(params.postLink, gp.actionURL, gp.header.CookieJar.ToString() /* gp.header.Location.Archive, */, "method=post")
	return response
}

func (gp *GlypeWebProxy) init(proxyURL string, timeout int, header *Header, userAgent string) bool {
	gp.SupportsPostRequest = false
	if gp.Initialised == true {
		return gp.SupportsPostRequest
	}
	gp.Transp.init("")
	gp.Initialised = true
	gp.Header = header
	gp.Header.init(proxyURL, userAgent)
	gp.Timeout = timeout
	var proxyPageContent string
	gp.LastLocation, _, proxyPageContent = GetRequest(proxyURL, gp.Transp.transport, gp.Timeout, gp.Header, 5)
	if strings.Contains(proxyPageContent, "ginf={url:") {
		gp.SupportsPostRequest = true //might be changed further down
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
			regi := regexp.MustCompile(`action="([^"]+)`)
			mts := regi.FindStringSubmatch(line)
			if mts != nil {
				action = mts[1]
				break forLoop
			}
		}
		//fmt.Println(action)
		//get all inputs fields names and values
		postData := allTextInputs[0][1] + "={url}&allowCookies=on"
		//fmt.Println(allInputs)

		gp.URL = proxyURL
		gp.PostData = postData
		gp.ActionURL = action
		if strings.Contains(gp.ActionURL, ".php") == false {
			gp.SupportsPostRequest = false
			return false
		}
		gp.resolveActionURL()
		gp.setRequisites(proxyURL, proxyPageContent)

	}
	if gp.SupportsPostRequest == true {
		return true
	}
	return false
}
