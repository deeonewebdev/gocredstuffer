package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"
)

var (
	baseURL       = &url.URL{Host: "https://api.anti-captcha.com/createTask", Scheme: "https", Path: "/"}
	checkInterval = 2 * time.Second
)

type Client struct {
	APIKey string
}

type antiCaptcha struct {
	twoCaptcha
	client     *Client
	ErrorCodes map[int]string
	Errors     map[int]string
}

func (ac *antiCaptcha) init(header *Header, apikey string, userAgent string) {
	ac.header = header
	ac.client = &Client{}
	ac.userAgent = userAgent
	ac.header.init("", ac.userAgent)
	ac.Name = "antiCaptcha"
	ac.ApiKey = apikey
	//set post method to application/json
	ac.header.setCustomRequestContentType("application/json")
	ac.ErrorCodes = make(map[int]string)
	ac.Errors = make(map[int]string)
	ac.ErrorCodes[1] = "ERROR_KEY_DOES_NOT_EXIST"
	ac.ErrorCodes[2] = "ERROR_ZERO_CAPTCHA_FILESIZE"
	ac.ErrorCodes[3] = "ERROR_TOO_BIG_CAPTCHA_FILESIZE"
	ac.ErrorCodes[4] = "ERROR_ZERO_BALANCE"
	ac.ErrorCodes[10] = "ERROR_IP_NOT_ALLOWED"
	ac.ErrorCodes[11] = "ERROR_CAPTCHA_UNSOLVABLE"
	ac.ErrorCodes[12] = "ERROR_BAD_DUPLICATES"
	ac.ErrorCodes[13] = "ERROR_NO_SUCH_METHOD"
	ac.ErrorCodes[14] = "ERROR_IMAGE_TYPE_NOT_SUPPORTED"
	ac.ErrorCodes[15] = "ERROR_NO_SUCH_CAPCHA_ID"
	ac.ErrorCodes[16] = "ERROR_EMPTY_COMMENT"
	ac.ErrorCodes[20] = "ERROR_IP_BLOCKED"
	ac.ErrorCodes[21] = "ERROR_TASK_ABSENT"
	ac.ErrorCodes[22] = "ERROR_TASK_NOT_SUPPORTED"
	ac.ErrorCodes[23] = "ERROR_INCORRECT_SESSION_DATA"
	ac.ErrorCodes[24] = "ERROR_PROXY_CONNECT_REFUSED"
	ac.ErrorCodes[25] = "ERROR_PROXY_CONNECT_TIMEOUT"
	ac.ErrorCodes[26] = "ERROR_PROXY_READ_TIMEOUT"
	ac.ErrorCodes[27] = "ERROR_PROXY_BANNED"
	ac.ErrorCodes[28] = "ERROR_PROXY_TRANSPARENT"
	ac.ErrorCodes[29] = "ERROR_RECAPTCHA_TIMEOUT"
	ac.ErrorCodes[30] = "ERROR_RECAPTCHA_INVALID_SITEKEY"
	ac.ErrorCodes[31] = "ERROR_RECAPTCHA_INVALID_DOMAIN"
	ac.ErrorCodes[32] = "ERROR_RECAPTCHA_OLD_BROWSER"
	ac.ErrorCodes[33] = "ERROR_TOKEN_EXPIRED"
	ac.ErrorCodes[34] = "ERROR_PROXY_HAS_NO_IMAGE_SUPPORT"
	ac.ErrorCodes[35] = "ERROR_PROXY_INCOMPATIBLE_HTTP_VERSION"
	ac.ErrorCodes[36] = "ERROR_FACTORY_SERVER_BAD_JSON"
	ac.ErrorCodes[37] = "ERROR_FACTORY_SERVER_ERRORID_MISSING"
	ac.ErrorCodes[38] = "ERROR_FACTORY_SERVER_ERRORID_NOT_ZERO"
	ac.ErrorCodes[39] = "ERROR_FACTORY_MISSING_PROPERTY"
	ac.ErrorCodes[40] = "ERROR_FACTORY_PROPERTY_INCORRECT_FORMAT"
	ac.ErrorCodes[41] = "ERROR_FACTORY_ACCESS_DENIED"
	ac.ErrorCodes[42] = "ERROR_FACTORY_PLATFORM_OPERATION_FAILED"
	ac.ErrorCodes[43] = "ERROR_FACTORY_PROTOCOL_BROKEN"
	ac.ErrorCodes[44] = "ERROR_FACTORY_TASK_NOT_FOUND"
	ac.ErrorCodes[45] = "ERROR_FACTORY_IS_SANDBOXED"
	ac.ErrorCodes[46] = "ERROR_PROXY_NOT_AUTHORISED"
	ac.ErrorCodes[47] = "ERROR_FUNCAPTCHA_NOT_ALLOWED"
	ac.ErrorCodes[48] = "ERROR_INVISIBLE_RECAPTCHA"
	ac.ErrorCodes[49] = "ERROR_FAILED_LOADING_WIDGET"
	ac.ErrorCodes[50] = "ERROR_VISIBLE_RECAPTCHA"
	ac.ErrorCodes[51] = "ERROR_ALL_WORKERS_FILTERED"
	ac.ErrorCodes[52] = "ERROR_NO_SLOT_AVAILABLE"
	ac.ErrorCodes[53] = "ERROR_FACTORY_SERVER_API_CONNECTION_FAILED"
	ac.ErrorCodes[54] = "ERROR_FACTORY_SERVER_OPERATION_FAILED"
	ac.Errors[52] = "stop"
	ac.Errors[53] = "stop"
	ac.Errors[54] = "stop"
	ac.ErrorCodes[45] = "stop"

}

func (ac *antiCaptcha) reportBad() {

}

func (ac *antiCaptcha) reportGood() {

}

func (ac *antiCaptcha) solve() {
	ac.client.APIKey = ac.ApiKey
	ac.twoCaptcha.solve()
}

func (ac *antiCaptcha) setSiteKey(siteKey string) {

}

func (ac *antiCaptcha) setCaptchaType(captchaType string) {

}

func (ac *antiCaptcha) setupHcaptchaV2(siteURL, proxyString string) {

}

func (ac *antiCaptcha) solveImageCaptcha() {
	data, err := ioutil.ReadFile(ac.filename)
	if err != nil {
		ac.Error = fmt.Errorf("error reading file")
		ac.completed = true
		ac.solveReturned <- &emptyStruct{}
		return
	}

	postBody := fmt.Sprintf(`{"clientKey":"%s","task": {"type": "ImageToTextTask","body":"%s"}}`,
		ac.ApiKey, base64.StdEncoding.EncodeToString(data))

	if ac.getCaptchaId(postBody) == false {
		ac.solveReturned <- &emptyStruct{}
		return
	}
	ac.sleep = 6
	//start polling for result
	ac.prepResultVars()
	ac.solveReturned <- &emptyStruct{}
}

func (ac *antiCaptcha) solveGoogleRecaptchaVersionTwo() {
	jsonString := `{"clientKey": "%s","task":{"type":"NoCaptchaTaskProxyless",`
	jsonString += `"websiteURL":"%s","websiteKey":"%s"}}`
	postBody := fmt.Sprintf(jsonString, ac.ApiKey, ac.pageURL, ac.pageRecaptchaId)

	if ac.getCaptchaId(postBody) == false {
		ac.solveReturned <- &emptyStruct{}
		return
	}
	//wait for 15 to 20 seconds
	ac.sleep = 18
	//poll for result
	//set the below vars on twocaptcha object and
	ac.prepResultVars()
	ac.solveReturned <- &emptyStruct{}
}

func (ac *antiCaptcha) solveGoogleRecaptchaVersionThree() {
	jsonString := `{"clientKey":"%s","task":{"type":"RecaptchaV3TaskProxyless",	"websiteURL":"%s",`
	jsonString += `"websiteKey":"%s","minScore": 0.3}}`
	postBody := fmt.Sprintf(jsonString, ac.ApiKey, ac.pageURL, ac.pageRecaptchaId)

	if ac.getCaptchaId(postBody) == false {
		ac.solveReturned <- &emptyStruct{}
		return
	}
	//wait for 15 to 20 seconds
	ac.sleep = 18
	//poll for result
	//set the below vars on twocaptcha object and
	ac.prepResultVars()
	ac.solveReturned <- &emptyStruct{}
}

func (ac *antiCaptcha) prepResultVars() {
	//set the below vars on twocaptcha object and
	ac.pp.postBody = fmt.Sprintf(`{"clientKey":"%s","taskId":%d}`, ac.ApiKey, ac.id)
	ac.pp.expecting = "string"
}

func (ac *antiCaptcha) decodeJSON(jsonString string) map[string]interface{} {
	//basic check
	jsonMap := make(map[string]interface{})
	json.Unmarshal([]byte(jsonString), &jsonMap)
	return jsonMap
}

func (ac *antiCaptcha) getCaptchaId(postBody string) bool {
	ac.nwp.init("", 20, ac.header, ac.userAgent)
	pollRez := ac.pollResponse(ac.InURL, postBody, "digit")
	var isInt bool
	ac.id, isInt = pollRez.(int)
	if isInt == false {
		ac.Error = fmt.Errorf("couldn't get an id from captcha service")
		ac.completed = true
		return false
	}
	return true
}

func (ac *antiCaptcha) pollResult() {
	//this will be called as a goroutine to poll for results
	time.Sleep(time.Second * time.Duration(ac.sleep))
	pollRez := ac.pollResponse(ac.OutURL, ac.pp.postBody, ac.pp.expecting)
	var isString bool
	ac.solution, isString = pollRez.(string)
	if isString == true {
		ac.Error = nil
	}
	ac.completed = true
	ac.pollResultReturned <- &emptyStruct{}
}

func (ac *antiCaptcha) pollResponse(requestURL, postBody string, expecting string) interface{} {
	var result interface{}
forLoop:
	for {
		fmt.Println("polling captcha result from " + ac.Name)
		responseJSON := ac.decodeJSON(ac.nwp.postRequest(requestURL, postBody, "", 20))
		//val, ok := errorMap[response]
		//respParts := strings.Split(response, "|")
		errorId, isInt := responseJSON["errorId"].(int)
		solutionMap, isMapOfStrings := responseJSON["solution"].(map[string]string)
		switch {
		case isInt == true && errorId == 0 &&
			expecting == "string" &&
			isMapOfStrings:
			result = solutionMap["text"]
			break forLoop
		case errorId == 0 && expecting == "digit":
			result = responseJSON["taskId"]
			break forLoop
		case isInt == false:
			//error  or result not set continue
		case isInt == true && ac.Errors[errorId] == "stop":
			//error we have to stop checking and panic
			fmt.Println("a fatal captcha error occurred: " + ac.ErrorCodes[errorId])
			ac.Error = fmt.Errorf("a fatal captcha error occurred: " + ac.ErrorCodes[errorId])
			return false //to make sure tc.error is not overwritten by calling function
		case isInt == true && ac.Errors[errorId] == "continue":
			//error but we can continue
			//repeat set of return to calling function
			//return and stop trying, combo will send
			//another request hopefully that one will fare better
			ac.Error = fmt.Errorf("continuing polling got: " + ac.ErrorCodes[errorId])
		default:
			//continue polling response was not understood
		}

		time.Sleep(time.Second * time.Duration(sleepInterval))
	}
	return result
}
