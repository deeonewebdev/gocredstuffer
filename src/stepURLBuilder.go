package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type StepURLBuilder struct {
	postBody
	url      []string
	content  *string
	cmb      *combo
	domain   string
	provider string
	username string
	email    string
	tld      string
	header   *Header
}

func (sub *StepURLBuilder) init(header *Header) *StepURLBuilder {
	sub.header = header
	sub.url = make([]string, 0)
	sub.content = nil
	return sub
}
func (sub *StepURLBuilder) resetURL() *StepURLBuilder {
	sub.url = make([]string, 0)
	return sub
}
func (sub *StepURLBuilder) fromCombo(cmb *combo) *StepURLBuilder {
	sub.cmb = cmb
	sub.email = cmb.username
	if strings.Contains(cmb.username, "@") {
		uparts := strings.Split(cmb.username, "@")
		sub.username = uparts[0]
		sub.domain = uparts[1]
		if strings.Contains(sub.domain, ".") {
			domParts := strings.Split(sub.domain, ".")
			sub.provider = domParts[0]
			sub.tld = domParts[1]

		}
	}
	return sub
}

func (sub *StepURLBuilder) fromContent(content *string) *StepURLBuilder {
	sub.content = content
	return sub
}

func (sub *StepURLBuilder) extrapolateVars(formString string) string {
	//user := sub.cmb.username

	formString = strings.Replace(formString, "{user}", url.QueryEscape(sub.username), -1)
	formString = strings.Replace(formString, "{email}", url.QueryEscape(sub.cmb.username), -1)
	formString = strings.Replace(formString, "{domain}", url.QueryEscape(sub.domain), -1)
	formString = strings.Replace(formString, "{provider}", url.QueryEscape(sub.provider), -1)
	formString = strings.Replace(formString, "{tld}", url.QueryEscape(sub.tld), -1)

	if strings.Contains(formString, "{Generate") {
		//match the generate string itself
		regString := `\{(Generate(?:LowString|CapLowString|CapString|NumLowString`
		regString += `|NumCapLowString|NumCapString|NumString|Md5String)_[0-9]{1,})\}`
		reg := regexp.MustCompile(regString)
		mts := reg.FindStringSubmatch(formString)
		if mts != nil {
			sach := fmt.Sprintf("{%s}", mts[1])
			mpats := strings.Split(mts[1], "_")
			repl := sub.call(mpats[0], CastToInteger(mpats[1]), "")
			formString = strings.Replace(formString, sach, url.QueryEscape(repl), -1)
		}
	}
	formString = strings.Replace(formString, "{pass}", url.QueryEscape(sub.cmb.password), -1)
	return formString
}

func (sub *StepURLBuilder) addURI(uri string) *StepURLBuilder {
	uri = sub.extrapolateVars(uri)
	switch {
	case strings.Contains(uri, "||+||") == false:
		sub.url = append(sub.url, uri)
		return sub
	}

	uriParts := strings.Split(uri, "||+||")
	for _, uriSegment := range uriParts {
		if strings.Contains(uriSegment, "::") == false {
			continue
		}
		subspl := strings.Split(uriSegment, "::")
		responsePart, regex := subspl[0], subspl[1]
		switch responsePart {
		case "content":
			sub.getRegexFromContent(regex)
		case "cookies":
			sub.getRegexFromCookies(regex)
		case "locations":
			sub.getRegexFromLocations(regex)
		case "reset":
			sub.init(sub.header)
		case "append":
			sub.url = append(sub.url, regex)
		}
	}
	return sub
}

func (sub *StepURLBuilder) getRegexFromContent(regex string) {
	content, err := url.QueryUnescape(*sub.content)
	if err != nil {
		content = *sub.content
	}
	rez, err := RegexpGetSubmatchFromString(regex, content)
	if err == nil && len(rez) > 0 {
		sub.url = append(sub.url, rez)
	}

}

func (sub *StepURLBuilder) getRegexFromCookies(regex string) {
	cookies := sub.header.CookieJar.ToString()
	rez, err := RegexpGetSubmatchFromString(regex, cookies)
	if err == nil {
		sub.url = append(sub.url, rez)
	}
}

func (sub *StepURLBuilder) getRegexFromLocations(regex string) {
	//TODO: handle glype web proxy case
	var locations []string
	var suc bool
	locations = sub.header.Location.Archive
	locations, suc = ReverseSlice(locations).([]string)
	if suc == true {
		locs := strings.Join(locations, "\n")
		rez, err := RegexpGetSubmatchFromString(regex, locs)
		if err == nil {
			sub.url = append(sub.url, rez)
		}
		return
	}
	panic("something horrible happened reverse slice did not return a slice of strings")
}

func (sub *StepURLBuilder) toString() string {
	//TODO: handle glype web proxy case
	//parse the urls list picking one from locations as a template
	location := sub.header.Location.Archive
	var lastLocation string
	if len(location) == 0 {
		lastLocation = sub.url[len(sub.url)-1]
	} else {
		lastLocation = location[len(location)-1]

	}
	llParse, err := url.Parse(lastLocation)
	if err != nil {
		panic("last location is not a valid url")
	}
	//uri := strings.Join(sub.url, "&")
	joiner := "?"
	nuri := ""
	for c, part := range sub.url {
		switch {
		case c == 0:
			joiner = ""
		case c > 1:
			fallthrough
		/* case strings.Contains(part, "://") == false:
		fallthrough */
		case strings.Contains(part, "?"):
			joiner = "&"
		default:
			joiner = "?"
		}
		nuri = nuri + joiner + part

	}
	resURL, err := llParse.Parse(nuri)
	if err != nil {
		panic(err.Error() + " parsing failed for " + nuri)
	}
	return resURL.String()
}
