package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CaptchaApiInterface interface {
	detectCaptcha(captchaTypeName string, stepRegex string, referenceToContent *string) bool
	setSiteKey(siteKey string)
	setCaptchaType(captchaType string)
	solve()
	setupImageCaptcha(filename string)
	setupRawImageCaptcha(base64ImageData string)
	setupGoogleRecaptchaV2(pageURL, proxyString string)
	setupHcaptchaV2(pageURL, proxyString string)
	setupGoogleRecaptchaV3(pageURL, proxyString string, minScore float32)
	getSolveReturned() chan *emptyStruct
	isEnterprise()
	pollResult()
	reportBad()
	reportGood()
	getCaptchaTaskId() int
	getName() string
	getPollResultReturned() chan *emptyStruct
	getError() error
	getSolution() string
	getApiKey() string
	init(header *Header, apikey string, userAgent string)
}

type CaptchaSolvingService struct {
	Name            string
	ApiKey          string
	InURL           string
	OutURL          string
	NumberOfWorkers int
	Channel         chan CaptchaApiInterface
	Api             CaptchaApiInterface
}

type CaptchaApis struct {
	CaptchaSolvers            []CaptchaSolvingService
	InitializedCaptchaSolvers []*CaptchaSolvingService // initialized
}

func (css *CaptchaSolvingService) closeChannel() {
	close(css.Channel)
}

func (css *CaptchaSolvingService) createChannel() {
	css.Channel = make(chan CaptchaApiInterface, 500)
}

func (css *CaptchaSolvingService) getApi() CaptchaApiInterface {
	return css.Api
}

func (css *CaptchaSolvingService) getChannel() chan CaptchaApiInterface {
	return css.Channel
}

func (ca *CaptchaApis) load(filename string) {
	//loads json
	apiStringList := ReadFile(filename)
	if len(apiStringList) > 0 && len(apiStringList[0]) > 10 {
		apiJSONString := strings.Join(apiStringList, "\n")
		err := json.Unmarshal([]byte(apiJSONString), &ca.CaptchaSolvers)
		if err != nil {
			panic(err)
		}
		//fmt.Println(apiJSONString + "\n")
	}

	//instantiates captcha workers
	for x := 0; x < len(ca.CaptchaSolvers); x++ {
		//we need an api key to use this
		if len(ca.CaptchaSolvers[x].ApiKey) == 0 {
			continue
		}
		switch ca.CaptchaSolvers[x].Name {
		case "twoCaptcha":
			ca.CaptchaSolvers[x].createChannel()
			//start 59 workers to call solve method of jobs
			ca.startCaptchaWorkers(ca.CaptchaSolvers[x].NumberOfWorkers, ca.CaptchaSolvers[x].Channel)
		case "antiCaptcha":
			ca.CaptchaSolvers[x].createChannel()
			//start 59 workers to call solve method of jobs
			ca.startCaptchaWorkers(ca.CaptchaSolvers[x].NumberOfWorkers, ca.CaptchaSolvers[x].Channel)
		default:
			panic("something's not right")
		}

	}

}

func (ca *CaptchaApis) startCaptchaWorkers(numberOfWorkers int, jobsChannel <-chan CaptchaApiInterface) {
	for x := 0; x < numberOfWorkers; x++ {
		go func(jobsChannel <-chan CaptchaApiInterface) {
			for job := range jobsChannel {
				//solve
				job.solve()
				//sleep three seconds to remain under the limit
				time.Sleep(time.Second * 3)
			}
		}(jobsChannel)
	}
}

func (ca *CaptchaApis) prep(hd *Header, userAgent string) {
	for x := 0; x < len(ca.CaptchaSolvers); x++ {
		//we need an api key to use this
		if len(ca.CaptchaSolvers[x].ApiKey) == 0 {
			//fmt.Println(captchaSolver.Name)
			continue
		}
		switch ca.CaptchaSolvers[x].Name {
		case "twoCaptcha":
			var tc twoCaptcha
			ca.CaptchaSolvers[x].Api = &tc
			//var hd Header
			tc.init(hd, ca.CaptchaSolvers[x].ApiKey, userAgent)
			tc.InURL = ca.CaptchaSolvers[x].InURL
			tc.OutURL = ca.CaptchaSolvers[x].OutURL
			ca.InitializedCaptchaSolvers = append(ca.InitializedCaptchaSolvers, &ca.CaptchaSolvers[x])
		case "antiCaptcha":
			var ac antiCaptcha
			ca.CaptchaSolvers[x].Api = &ac
			//var hd Header
			ac.init(hd, ca.CaptchaSolvers[x].ApiKey, userAgent)
			ac.InURL = ca.CaptchaSolvers[x].InURL
			ac.OutURL = ca.CaptchaSolvers[x].OutURL
			ca.InitializedCaptchaSolvers = append(ca.InitializedCaptchaSolvers, &ca.CaptchaSolvers[x])
		default:
			panic("we should never get here")
		}
	}
}

func (ca *CaptchaApis) pick(id int, hd *Header, userAgent string) (CaptchaApiInterface, chan CaptchaApiInterface) {
	//we need new instances each time pick is called
	//because we can't share an instance of a captcha api
	ca.prep(hd, userAgent) //done
	if len(ca.InitializedCaptchaSolvers) < 1 {
		panic("you have not specified any 2captcha or anticaptcha api key")
	}

	ind := id % len(ca.InitializedCaptchaSolvers)
	captchaSolver := ca.InitializedCaptchaSolvers[ind]
	//fmt.Println(captchaSolver.getApi(), captchaSolver.getChannel())
	return captchaSolver.getApi(), captchaSolver.getChannel()
}

func (ca *CaptchaApis) closeChannels() {
	//this might cause panic try to recover
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	for _, captchaSolver := range ca.InitializedCaptchaSolvers {
		captchaSolver.closeChannel()
	}
}
