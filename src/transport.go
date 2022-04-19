package main

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
)

type Transport struct {
	id        int
	transport *http.Transport
}

func (tp *Transport) init(proxyURL string) bool {
	proxyURL = strings.ToLower(proxyURL)
	var tr *http.Transport
	switch true {
	case strings.Contains(proxyURL, "socks"):
		proxyParse, _ := url.Parse(proxyURL)
		/* proxyParts := strings.Split(strings.Replace(proxyURL, "//", "", -1), ":")
		proxyHostPort := fmt.Sprintf("%s:%s", proxyParts[1], proxyParts[2])
		//fmt.Printf("%s                    \n", proxyHostPort)
		dialer, err := proxy.SOCKS5("tcp", proxyHostPort, nil, proxy.Direct)
		if err != nil {
			//fmt.Printf("error: %s\n", err.Error())
			return false
			//os.Exit(1)
		} */
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyParse),
		}
	case proxyURL == "":
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	default:
		pURL, _ := url.Parse(proxyURL)
		tr = &http.Transport{
			Proxy:           http.ProxyURL(pURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	tp.transport = tr
	return true
}
