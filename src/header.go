package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Header struct {
	Request       *Request
	Response      Response
	Location      Location
	CookieJar     *MyCookie
	Authorization []string
	Initialised   bool
}

type Request struct {
	Methods           []string
	KeyValue          map[string]string
	CustomContentType string
	HeadersArchive    []string //for debugging purposes
}

type Response struct {
	HeadersArchive []string
}

type Location struct {
	Archive      []string
	RequestChain []string
}

func (rq *Request) setValue(key, value string) {
	rq.KeyValue[key] = value
}

func (hd *Header) init(requestURL string, userAgent string) {
	if hd.Initialised {
		//this method should ideally be called only once
		return
	}
	hd.Request = &Request{}
	hd.Request.KeyValue = make(map[string]string)
	hd.Location.RequestChain = make([]string, 0)
	hd.Initialised = true
	//let the library handle next line
	//hd.Request.KeyValue["Referer"] = "" //full link address
	//hd.Request.KeyValue["Accept-Encoding"] = "deflate"
	hd.Request.KeyValue["Cache-Control"] = "no-cache"
	hd.Request.KeyValue["Pragma"] = "no-cache"
	hd.Request.KeyValue["Upgrade-Insecure-Requests"] = "1"
	//hd.Request.KeyValue["Cookie"] = ""
	hd.Request.KeyValue["Connection"] = "keep-alive"
	hd.Request.KeyValue["Accept-Language"] = "en-US,en;q=0.5"
	hd.Request.KeyValue["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
	/* hd.Request.KeyValue["Sec-Fetch-Dest"] = "document"
	hd.Request.KeyValue["Sec-Fetch-Mode"] = "navigate"
	hd.Request.KeyValue["Sec-Fetch-Site"] = "none"
	hd.Request.KeyValue["Sec-Fetch-User"] = "?1" */
	hd.Request.KeyValue["User-Agent"] = userAgent

	hd.Location.Archive = append(hd.Location.Archive, requestURL)
	var cookieJar MyCookie
	hd.CookieJar = &cookieJar
	//hd.Request.KeyValue["Origin"] = ""  //host, only sent during post requests

}

func (hd *Header) setAuthorization(value []string) {
	hd.Authorization = value
}

func (hd *Header) removeAuthorization(value string) {
	hd.Authorization = []string{}
}

func (hd *Header) LastLocation() string {
	return hd.Location.Archive[len(hd.Location.Archive)-1]
}

func (hd *Header) requestToString() string {
	header := ""
	//if request.Method is post add an origin header automatically
	//if Request Method is get remove it
	for key, value := range hd.Request.KeyValue {
		header += fmt.Sprintf("%s: %s\n", key, value)
	}
	return "\n\n" + header
}

func (hd *Header) archiveCurrentRequestHeaders() {
	header := ""
	//if request.Method is post add an origin header automatically
	//if Request Method is get remove it
	for key, value := range hd.Request.KeyValue {
		if len(value) > 0 {
			header += fmt.Sprintf("%s: %s\n", key, value)
		}
	}
	hd.Request.HeadersArchive = append(hd.Request.HeadersArchive, header)
}

func (hd *Header) responseToString() string {
	return "\n" + strings.Join(hd.Response.HeadersArchive, "\n\n\n")
}

func (hd *Header) setUserAgent(agent string) {
	hd.Request.KeyValue["User-Agent"] = agent
}

func (hd *Header) setCustomValue(key, value string) {
	hd.Request.KeyValue[key] = value
	if key == "Referer" {
		hd.Location.Archive = append(hd.Location.Archive, value)
	}
}

func (hd *Header) removeCustomValue(key string) {
	delete(hd.Request.KeyValue, key)
}

func (hd *Header) setCustomRequestContentType(contentType string) {
	if _, ok := hd.Request.KeyValue["content-type"]; ok {
		//stop dont wanna override what's set in config
		return
	}
	hd.Request.CustomContentType = contentType
	hd.Request.KeyValue["Content-Type"] = contentType
}

func (hd *Header) setRequestHeaders(requestURL string, req *http.Request, method string) {
	//set previous location as referer in request
	//if method is post set origin as scheme://host of last location
	//add this location to Location archive
	hd.Location.RequestChain = append(hd.Location.RequestChain, method+"=>"+requestURL)
	hd.Location.Archive = append(hd.Location.Archive, requestURL)
	hd.Request.Methods = append(hd.Request.Methods, method) //for debug purposes
	lastLocation := ""
	if len(hd.Location.Archive) > 1 {
		lastLocation = hd.Location.Archive[len(hd.Location.Archive)-2]
		if len(lastLocation) > 5 {

			hd.Request.KeyValue["Referer"] = lastLocation
		}
	}

	if len(lastLocation) > 1 &&
		strings.Contains(lastLocation, "http") &&
		method == "post" {
		/* hd.Request.KeyValue["Content-Type"] = hd.Request.CustomContentType
		if len(hd.Request.CustomContentType) == 0 {
			hd.Request.KeyValue["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
		} */
		locationParse, _ := url.Parse(lastLocation)
		origin := fmt.Sprintf("%s://%s", locationParse.Scheme, locationParse.Host)
		hd.Request.KeyValue["Origin"] = origin
	}

	delete(req.Header, "X-Https")
	for key, value := range hd.Request.KeyValue {
		if len(value) > 0 {
			//req.Header.Add(key, value)
			req.Header[key] = []string{value}

		}
	}
	//req.Header["Accept-Encoding"] = []string{"deflate"}
	//delete(req.Header, "Accept-Encoding")

	//req.Header["Accept-Encoding"] = []string{"deflate"}
	if requestURL == "https://signin.sso.members1st.org/api/v1/authn" && method != "option" {
		//req.Header["content-type"] = []string{"application/json"}
		//req.Header["x-okta-user-agent-extended"] = []string{"okta-signin-widget-3.1.0"}
		//fmt.Println("popopopopopopo", req.Header)
	}

	//req.Header["Accept-Encoding"] = []string{"gzip", "deflate"}
}

func (hd *Header) getHeaderValue(hkey string) string {
	var retVal string
	if hValue, ok := hd.Request.KeyValue[hkey]; ok {
		retVal = hValue
	}
	return retVal
}

func (hd *Header) saveHeaders(status string, headers http.Header) {
	//read all header values to a string and add it to response headers list
	//also update request headers accordingly
	//might even handle cookies saving here
	//header is a map[string][]string

	var responseHeaderString string
	for key, value := range headers {

		switch strings.ToLower(key) {
		case "location":
			//add this location to Location archive
			location := value[len(value)-1] //only interested in last element
			urlParse, _ := url.Parse(hd.LastLocation())
			locParse, err := urlParse.Parse(location)
			if err == nil {
				//hd.Location.Archive = append(hd.Location.Archive, locParse.String())
				hd.Location.RequestChain = append(hd.Location.RequestChain, "location-redirect=>"+locParse.String())
			}
		case "set-cookie":
			//this will affect request headers
			hd.CookieJar.Add(ExtractCookieFromHeaders(headers)) //defined in functions.go
			hd.Request.KeyValue["Cookie"] = hd.CookieJar.ToString()
		}

		//add entry in response header string
		for _, v := range value {
			responseHeaderString += fmt.Sprintf("%s: %s\n", key, v)
		}
	}
	hd.Response.HeadersArchive = append(hd.Response.HeadersArchive, status+"\n"+responseHeaderString)
	//save a copy of current request headers in archive
	hd.archiveCurrentRequestHeaders()
}
