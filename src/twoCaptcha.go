package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var _inErrors = map[string]string{
	"ERROR_WRONG_USER_KEY":           "stop",
	"ERROR_KEY_DOES_NOT_EXIST":       "stop",
	"ERROR_ZERO_BALANCE":             "stop",
	"ERROR_pageURL":                  "stop",
	"ERROR_NO_SLOT_AVAILABLE":        "stop",
	"ERROR_TOO_BIG_CAPTCHA_FILESIZE": "stop",
	"ERROR_WRONG_FILE_EXTENSION":     "stop",
	"ERROR_IMAGE_TYPE_NOT_SUPPORTED": "stop",
	"ERROR_UPLOAD":                   "stop",
	"ERROR_IP_NOT_ALLOWED":           "stop",
	"IP_BANNED":                      "stop",
	"ERROR_BAD_TOKEN_OR_pageURL":     "stop",
	"ERROR_GOOGLEKEY":                "stop",
	"ERROR_CAPTCHAIMAGE_BLOCKED":     "stop",
	"TOO_MANY_BAD_IMAGES":            "stop",
	"ERROR: ":                        "stop",
	"ERROR_BAD_PARAMETERS":           "stop",
	"ERROR_BAD_PROXY":                "stop",
	"MAX_USER_TURN":                  "continue",
	"ERROR_ZERO_CAPTCHA_FILESIZE":    "continue",
}

var _resErrors = map[string]string{
	"ERROR_CAPTCHA_UNSOLVABLE":      "stop",
	"ERROR_WRONG_USER_KEY":          "stop",
	"ERROR_KEY_DOES_NOT_EXIST":      "stop",
	"ERROR_WRONG_ID_FORMAT":         "stop",
	"ERROR_WRONG_CAPTCHA_ID":        "stop",
	"ERROR_BAD_DUPLICATES":          "stop",
	"REPORT_NOT_RECORDED":           "stop",
	"ERROR_DUPLICATE_REPORT":        "stop",
	"ERROR: ":                       "stop",
	"ERROR_IP_ADDRES":               "stop",
	"ERROR_TOKEN_EXPIRED":           "stop",
	"ERROR_EMPTY_ACTION":            "stop",
	"ERROR_PROXY_CONNECTION_FAILED": "stop",
	"CAPCHA_NOT_READY":              "continue",
}

type twoCaptcha struct {
	id                  int
	solution            string
	Error               error
	header              *Header
	nwp                 *noWebProxy
	ApiKey              string
	InURL               string
	OutURL              string
	pageURL             string
	captchaType         string
	isEnterpriseCaptcha bool
	filename            string
	base64ImageData     string
	pageRecaptchaId     string
	proxyString         string
	Name                string
	minScore            float32
	completed           bool
	sleep               int
	solveReturned       chan *emptyStruct
	pollResultReturned  chan *emptyStruct
	pp                  pollParam
	requestMethod       string
	userAgent           string
}

type pollParam struct {
	postBody  string
	expecting string
}

type emptyStruct struct{}

const sleepInterval int = 5

//can only havfunc (ac *antiCaptcha)e maximum of 59 workers
func (tc *twoCaptcha) init(header *Header, apikey string, userAgent string) {

	var nwp noWebProxy
	tc.header = header
	tc.nwp = &nwp
	tc.ApiKey = apikey
	tc.userAgent = userAgent
	tc.Name = "twoCaptcha"
	tc.solveReturned = make(chan *emptyStruct, 1)
	tc.pollResultReturned = make(chan *emptyStruct, 1)
	tc.requestMethod = "GET"

}

func (tc *twoCaptcha) isEnterprise() {
	tc.isEnterpriseCaptcha = true
}

func (tc *twoCaptcha) detectCaptcha(captchaType string, captchaRegex string, content *string) bool {
	switch strings.ToLower(captchaType) {
	case "googlecaptchav2":
		reg := regexp.MustCompile(captchaRegex)
		mts := reg.FindStringSubmatch(*content)
		if mts != nil {
			tc.pageRecaptchaId = mts[1]
			return true
		}
	}
	return false
}

func (tc *twoCaptcha) setSiteKey(siteKey string) {
	tc.pageRecaptchaId = siteKey
}

func (tc *twoCaptcha) setCaptchaType(captchaType string) {
	tc.captchaType = captchaType
}

func (tc *twoCaptcha) setupImageCaptcha(filename string) {
	tc.captchaType = "imagecaptcha"
	tc.filename = filename
}

func (tc *twoCaptcha) setupRawImageCaptcha(base64ImageData string) {
	tc.captchaType = "imagecaptcha"
	tc.filename = ""
	tc.base64ImageData = base64ImageData
}

func (tc *twoCaptcha) setupHcaptchaV2(pageURL, proxyString string) {
	tc.captchaType = "hcaptchav2"
	tc.pageURL = url.QueryEscape(pageURL)
	tc.proxyString = proxyString
}

func (tc *twoCaptcha) setupGoogleRecaptchaV2(pageURL, proxyString string) {
	tc.captchaType = "googlecaptchav2"
	tc.pageURL = url.QueryEscape(pageURL)
	tc.proxyString = proxyString
}

func (tc *twoCaptcha) getName() string {
	return tc.Name
}

func (tc *twoCaptcha) getCaptchaTaskId() int {
	return tc.id
}

func (tc *twoCaptcha) setupGoogleRecaptchaV3(pageURL, proxyString string, minScore float32) {
	tc.captchaType = "googlecaptchav3"
	tc.pageURL = url.QueryEscape(pageURL)
	tc.proxyString = proxyString
	tc.minScore = minScore
}

func (tc *twoCaptcha) getSolveReturned() chan *emptyStruct {
	return tc.solveReturned
}

func (tc *twoCaptcha) getPollResultReturned() chan *emptyStruct {
	return tc.pollResultReturned
}

func (tc *twoCaptcha) getApiKey() string {
	return tc.ApiKey
}

func (tc *twoCaptcha) getError() error {
	return tc.Error
}

func (tc *twoCaptcha) getSolution() string {
	return tc.solution
}

func (tc *twoCaptcha) isDone() bool {
	return tc.completed
}

func (tc *twoCaptcha) solve() {
	switch strings.ToLower(tc.captchaType) {
	case "imagecaptcha":
		tc.solveImageCaptcha()
	case "hcaptchav2":
		fallthrough
	case "googlecaptchav2":
		tc.solveGoogleRecaptchaVersionTwo()
	case "googlecaptchav3":
		tc.solveGoogleRecaptchaVersionThree()
	default:
		tc.completed = true
	}
}

func (tc *twoCaptcha) reportBad() {
	fmt.Println("reporting captcha solution as bad")
	postBody := fmt.Sprintf("key=%s&id=%d&action=reportbad", tc.ApiKey, tc.id)
	tc.requestMethod = "GET"
	tc.pollResponse(tc.OutURL, postBody, _resErrors, "string")
}

func (tc *twoCaptcha) reportGood() {
	//fmt.Println("reporting captcha solution as good")
	postBody := fmt.Sprintf("key=%s&id=%d&action=reportgood", tc.ApiKey, tc.id)
	tc.requestMethod = "GET"
	tc.pollResponse(tc.OutURL, postBody, _resErrors, "string")
}

func (tc *twoCaptcha) solveImageCaptcha() {
	//attempt to read image file and base64 encode it

	var rawImageData string
	//tc.header.setCustomRequestContentType("application/x-www-form-urlencoded; charset=UTF-8")
	if len(tc.filename) == 0 {
		rawImageData = tc.base64ImageData
		if len(tc.base64ImageData) < 5 {
			tc.completed = true
			tc.solveReturned <- &emptyStruct{}
			return
		}
	} else {
		data, err := ioutil.ReadFile(tc.filename)
		if err != nil {
			tc.Error = fmt.Errorf("error reading file")
			tc.completed = true
			tc.solveReturned <- &emptyStruct{}
			return
		}
		rawImageData = base64.StdEncoding.EncodeToString(data)
	}

	postBody := fmt.Sprintf("key=%s&method=base64&body=%s&lang=en",
		tc.ApiKey, rawImageData)

	if tc.getCaptchaId(postBody, "POST") == false {
		tc.solveReturned <- &emptyStruct{}
		return
	}
	tc.sleep = 6
	//start polling for result
	tc.prepResultVars()
	tc.solveReturned <- &emptyStruct{}
}

func (tc *twoCaptcha) prepResultVars() {
	//set the below vars on twocaptcha object and
	tc.pp.postBody = fmt.Sprintf("key=%s&id=%d&action=get", tc.ApiKey, tc.id)
	tc.pp.expecting = "string"
}

func (tc *twoCaptcha) getCaptchaId(postBody string, requestMethod string) bool {
	tc.nwp.init("", 20, tc.header, tc.userAgent)
	//this uses get method not post method
	oldRequestMethod := tc.requestMethod
	tc.requestMethod = requestMethod
	pollRez := tc.pollResponse(tc.InURL, postBody, _inErrors, "digit")
	tc.requestMethod = oldRequestMethod
	var isInt bool
	tc.id, isInt = pollRez.(int)
	if isInt == false || tc.id == 0 {
		tc.Error = fmt.Errorf("couldn't get an id from captcha service")
		tc.completed = true
		return false
	}
	fmt.Printf("got twocaptcha task id: %d\n", tc.id)
	return true
}

func (tc *twoCaptcha) solveGoogleRecaptchaVersionTwo() {

	postBody := fmt.Sprintf("key=%s&method=userrecaptcha&googlekey=%s&pageURL=%s",
		tc.ApiKey, tc.pageRecaptchaId, tc.pageURL)

	if tc.isEnterpriseCaptcha {
		postBody += "&enterprise=1"
	}
	if strings.ToLower(tc.captchaType) == "hcaptchav2" {
		postBody = fmt.Sprintf("key=%s&method=hcaptcha&sitekey=%s&pageURL=%s",
			tc.ApiKey, tc.pageRecaptchaId, tc.pageURL)
	}
	if tc.getCaptchaId(postBody, "GET") == false {
		tc.solveReturned <- &emptyStruct{}
		return
	}
	//wait for 15 to 20 seconds
	tc.sleep = 18
	//poll for result
	//set the below vars on twocaptcha object and
	tc.prepResultVars()
	tc.solveReturned <- &emptyStruct{}
}

func (tc *twoCaptcha) solveGoogleRecaptchaVersionThree() {

	postBody := fmt.Sprintf("key=%s&method=userrecaptcha&version=v3&min_score=%f&googlekey=%s&pageurl=%s",
		tc.ApiKey, tc.minScore, tc.pageRecaptchaId, tc.pageURL)

	if tc.isEnterpriseCaptcha {
		postBody += "&enterprise=1"
	}

	if tc.getCaptchaId(postBody, "GET") == false {
		tc.solveReturned <- &emptyStruct{}
		return
	}
	//wait for 15 to 20 seconds
	tc.sleep = 18
	//poll for result
	//set the below vars on twocaptcha object and
	tc.prepResultVars()
	tc.solveReturned <- &emptyStruct{}
}

func (tc *twoCaptcha) pollResult() {
	//this will be called as a goroutine to poll for results
	time.Sleep(time.Second * time.Duration(tc.sleep))
	pollRez := tc.pollResponse(tc.OutURL, tc.pp.postBody, _resErrors, tc.pp.expecting)
	var isString bool
	tc.solution, isString = pollRez.(string)
	if isString == true {
		tc.Error = nil
	}
	tc.completed = true
	tc.pollResultReturned <- &emptyStruct{}
	fmt.Printf("got captcha solution from twocaptcha length: %d\n", len(tc.solution))
}

func (tc *twoCaptcha) pollResponse(requestURL, postBody string, errorMap map[string]string, expecting string) interface{} {
	var result interface{}
forLoop:
	for {
		fmt.Println("polling captcha result from " + tc.Name)
		//this might be get or post
		var response string
		if tc.requestMethod == "GET" {
			response = tc.nwp.getRequest(fmt.Sprintf("%s?%s", requestURL, postBody), "", 20)
		} else {
			tc.nwp.getHeader().setCustomRequestContentType("application/x-www-form-urlencoded; charset=UTF-8")
			response = tc.nwp.postRequest(requestURL, postBody, "", 20)
		}
		//fmt.Println("[[[[[[[[[", response, "]]]]]]]]]]]]]]]]]")
		val, ok := errorMap[response]
		//fmt.Println(response)
		respParts := strings.Split(response, "|")
		switch {
		case strings.Contains(response, "OK|") && expecting == "string":
			result = respParts[1]
			break forLoop
		case strings.Contains(response, "OK|") && expecting == "digit":
			result, _ = strconv.Atoi(respParts[1])
			break forLoop
		case response == "OK_REPORT_RECORDED":
			return true
		case ok == false:
			//error  or result not set continue
			fmt.Printf("got: %s continuing\n", response)
		case val == "continue":
			//error but we can continue
			//repeat set of return to calling function
			//return and stop trying, combo will send
			//another request hopefully that one will fare better
			tc.Error = fmt.Errorf("continuing polling got: " + response)
		case val == "stop":
			//error we have to stop checking and panic
			fmt.Println("a fatal captcha error occurred: " + response)
			tc.Error = fmt.Errorf("a fatal captcha error occurred: " + response)
			return false //to make sure tc.error is not overwritten by calling function
		}

		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
	return result
}
