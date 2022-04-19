package main

import "fmt"

type WebProxyInit struct {
	wpInterface WebProxyInterface
	Glype       GlypeWebProxy
	PhpProxy    PhpWebProxy
}

func (wpi *WebProxyInit) init(requestURL string, timeout int, userAgent string) WebProxyInterface {
	var hd1, hd2 Header
	switch {
	case wpi.Glype.init(requestURL, timeout, &hd1, userAgent) == true:
		return &wpi.Glype
	default:
		fmt.Println("\n", requestURL, "--------------")
		wpi.PhpProxy.init(requestURL, timeout, &hd2, userAgent)
		return &wpi.PhpProxy
	}
}
