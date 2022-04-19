package main

import (
	"fmt"
	"regexp"
	"strings"
)

type MyCookie struct {
	Cookies   string
	cookieMap map[string]string
	nodelete  string
}

func (c *MyCookie) Add(cookieString string) {
	//split string
	cookieParts := strings.Split(cookieString, ";")
	if strings.Contains(cookieString, "nodelete::") {
		parts := strings.Split(cookieString, "::")
		c.nodelete += parts[1]
		return
	}
	for _, acookie := range cookieParts {
		//split by =
		kv := strings.Split(acookie, "=")
		if c.cookieMap == nil {
			c.cookieMap = make(map[string]string)
		}

		switch {
		case len(kv) > 2:
			//weird cookie value! set the remaining list item as value of first
			c.cookieMap[strings.TrimSpace(kv[0])] = strings.Join(kv[1:], "=")
		case len(kv) == 2 && len(kv[1]) > 0:
			c.cookieMap[strings.TrimSpace(kv[0])] = kv[1]
		case len(kv) == 2 && len(kv[1]) == 0:
			//delete mode we are required to delete cookie
			if !strings.Contains(c.nodelete, kv[0]) {
				c.Delete(kv[0])
			}
		}
	}
}

func (c *MyCookie) replaceIn(templateStr string) string {
	regie := regexp.MustCompile(`\{cookie::([^}]+)\}`)
	matches := regie.FindAllStringSubmatch(templateStr, -1)
	var search, replace string
	var ok bool
	if matches != nil {
		for _, matchitem := range matches {
			search = matchitem[0]
			if replace, ok = c.cookieMap[matchitem[1]]; ok {
				//do something here
			} else {
				replace = ""
			}
			templateStr = strings.ReplaceAll(templateStr, search, replace)
		}
	}
	return templateStr
}

func (c *MyCookie) Delete(cookieKey string) {
	delete(c.cookieMap, cookieKey)
}

func (c *MyCookie) ToString() string {
	c.Cookies = ""
	for k, v := range c.cookieMap {
		c.Cookies += fmt.Sprintf("%s=%s; ", k, v)
	}
	return c.Cookies
}
