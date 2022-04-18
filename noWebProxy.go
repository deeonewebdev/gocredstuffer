package main

type noWebProxy struct {
	SessionCookie MyCookie
	Header        *Header
	Timeout       int
	Initialised   bool
	Transp        Transport
}

func (wp *noWebProxy) init(proxyURL string, timeout int, header *Header, userAgent string) {
	wp.Transp.init(proxyURL)
	if wp.Initialised == true {
		return
	}
	wp.Header = header
	wp.Initialised = true
	wp.Header.init("", userAgent)
	wp.Timeout = timeout
}

func (wp *noWebProxy) getRequest(requestURL string, proxy string, timeout int) string {
	//use cookie saved in object property
	//then send post request to proxy to fetch the page

	var response string
	_, _, response = GetRequest(requestURL, wp.Transp.transport, timeout, wp.Header, 5)
	return response
}

func (wp *noWebProxy) optionRequest(requestURL string, proxy string, timeout int) string {
	//use cookie saved in object property
	//then send post request to proxy to fetch the page

	var response string
	_, _, response = OptionRequest(requestURL, wp.Transp.transport, timeout, wp.Header, 5)
	return response
}

func (wp *noWebProxy) saveCookies(cookie string) {
	/* if len(cookie) > 3 {
		wp.sessionCookie = cookie
	} */
}

func (wp *noWebProxy) getHeader() *Header {
	return wp.Header
}

func (wp *noWebProxy) postRequest(requestURL string, postBody string, proxy string, timeout int) string {
	//use cookie saved in object property
	//then send post request to proxy
	//include postmethod parameter in postdata
	var params = postParameters{
		postLink: requestURL,
		postBody: postBody,
		transp:   wp.Transp.transport,
		timeout:  timeout,
		header:   wp.Header,
	}
	var response, cookie string
	_, cookie, response = PostRequest(params, 5)
	wp.saveCookies(cookie)

	return response
}
