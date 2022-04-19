package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type postBody struct {
	Inputs              url.Values
	jsonPostData        string
	content             string
	cookieJar           *MyCookie
	username            string
	password            string
	previouslyGenerated []string
	URL                 *url.URL
}

func (pb *postBody) init() *postBody {
	pb.Inputs = url.Values{}
	return pb
}

func (pb *postBody) fromURL(URL string) *postBody {
	var err error
	pb.URL, err = url.Parse(URL)
	if err != nil {
		panic("post body couldn't parse url")
	}
	return pb
}

func (pb *postBody) fromContent(content string) *postBody {
	cont, err := url.QueryUnescape(content)
	if err == nil {
		cont = content
	}
	cont = html.UnescapeString(content)

	pb.content = cont
	return pb
}

//possible problems here
func (pb *postBody) fromCombo(cmb *combo) *postBody {
	switch {
	case cmb.wp != nil && cmb.wp.isInitialised() == true:
		pb.cookieJar = cmb.wp.getHeader().CookieJar
	case cmb.nwp.Initialised == true:
		pb.cookieJar = cmb.nwp.Header.CookieJar
	}
	pb.password = cmb.password
	pb.username = cmb.username
	return pb
}

func (pb *postBody) getPostVars(steps []StepToGetPostVar) *postBody {
	//safegaurd against invalid steps
	if len(steps) == 0 || steps == nil {
		return pb
	}

	for _, postVarStep := range steps {
		//loop through content lines until we find match
		pb.lineWith(postVarStep)
		pb.allInputsFromTo(postVarStep)
		pb.cookieName(postVarStep)
		pb.Reset(postVarStep)
	}
	//fmt.Println(pb.Inputs, "########")
	return pb
}

func (pb *postBody) replaceIn(formString string) string {
	regie := regexp.MustCompile(`\{([a-zA-Z0-9_.-]+)\}`)
	mtses := regie.FindAllStringSubmatch(formString, -1)
	if mtses != nil {
		for _, mts := range mtses {
			search := fmt.Sprintf("{%s}", mts[1])
			formString = strings.Replace(formString, search, pb.Inputs.Get(mts[1]), -1)
		}
	}
	return formString
}

func (pb *postBody) extrapolateVars(formString string) string {
	user := pb.username
	var dt datetime
	if strings.Contains(pb.username, "@") {
		parts := strings.Split(pb.username, "@")
		if len(parts[1]) > 4 && strings.Contains(parts[1], ".") {
			//this must be a valid email
			//get the user portion
			user = parts[0]
		}
	}
	formString = strings.Replace(formString, "{user}", user, -1)
	formString = strings.Replace(formString, "{email}", pb.username, -1)
	//replace previous contentb64 var
	formString = strings.Replace(formString, "{contentB64}", base64.StdEncoding.EncodeToString([]byte(pb.content)), -1)
	formString = dt.replaceIn(formString)
	formString = pb.cookieJar.replaceIn(formString)

	if strings.Contains(formString, "{Generate") {
		//match the generate string itself
		regString := `\{(Generate(?:LowString|CapLowString|CapString|NumLowString`
		regString += `|NumCapLowString|NumCapString|NumString|Md5String|PrevString|Template|TemplateHex)_[^}]{1,})\}`
		reg := regexp.MustCompile(regString)
		//mts := reg.FindStringSubmatch(formString)
		mtses := reg.FindAllStringSubmatch(formString, -1)
		if mtses != nil {
			for _, mts := range mtses {
				sach := fmt.Sprintf("{%s}", mts[1])
				mpats := strings.Split(mts[1], "_")
				var repl string
				if strings.Contains(mpats[0], "GenerateTemplate") || strings.Contains(mpats[0], "GenerateTemplateHex") {
					repl = pb.call(mpats[0], -1, mpats[1])
				} else {
					repl = pb.call(mpats[0], CastToInteger(mpats[1]), "")

				}
				formString = strings.Replace(formString, sach, repl, -1)
			}
		}
	}

	formString = strings.Replace(formString, "{pass}", pb.password, -1)
	formString = pb.replaceIn(formString)
	formString = pb.transform(formString)
	return formString
}

func (pb *postBody) transform(item string) string {
	//user regex to find all parts that require transformation
	//e.g base64::((string_to_be_transformed))
	regie := regexp.MustCompile(`(base64|base64_decode|url_encode|url_decode|md5sum)::\(\((.+?)\)\)`)
	mts := regie.FindAllStringSubmatch(item, -1)
	if mts != nil {
		for _, mt := range mts {
			search := mt[0]
			repl := pb.getTransformation(mt[1], mt[2])
			item = strings.Replace(item, search, repl, 1) //replace only once in case there are other matches
		}
	}
	return item
}

func (pb *postBody) getTransformation(transFunc, toTransform string) string {

	switch transFunc {
	case "base64":
		toTransform = base64.StdEncoding.EncodeToString([]byte(toTransform))
	case "base64_decode":
		val, err := base64.StdEncoding.DecodeString(toTransform)
		toTransform = ""
		if err != nil {
			toTransform = string(val)
		}
	case "url_encode":
		toTransform = pb.urlEncode(toTransform)
	case "url_decode":
		toTransform = pb.urlDecode(toTransform)
	case "md5sum":
		toTransform = fmt.Sprintf("%x", md5.Sum([]byte(toTransform)))
	}

	return toTransform
}

func (pb *postBody) mergeWith(formString string) *postBody {
	//also replaces placeholders with values?
	//check for json post method
	if strings.Contains(formString, "json::") {
		pb.jsonPostData = pb.extrapolateVars(formString)
		return pb
	}
	formStringParts := strings.Split(formString, "&")
	for _, input := range formStringParts {
		inpParts := strings.Split(input, "=")
		if len(inpParts) == 2 {
			//this will overwrite any preexisting inputs
			valu := inpParts[1]
			if strings.Contains(inpParts[1], "{") {
				valu = pb.extrapolateVars(inpParts[1])
			}
			pb.Inputs.Add(inpParts[0], valu)
		}
	}
	return pb
}

func (pb *postBody) call(funcName string, length int, template string) string {
	var retVal string
	var gn Generator
	gn.init()
	switch funcName {
	case "GenerateLowString":
		retVal = gn.generateLowString(length)
	case "GenerateCapLowString":
		retVal = gn.generateCapLowString(length)
	case "GenerateCapString":
		retVal = gn.generateCapString(length)
	case "GenerateNumLowString":
		retVal = gn.generateNumLowString(length)
	case "GenerateNumCapLowString":
		retVal = gn.generateNumCapLowString(length)
	case "GenerateNumCapString":
		retVal = gn.generateNumCapString(length)
	case "GenerateNumString":
		retVal = gn.generateNumString(length)
	case "GenerateMd5String":
		retVal = gn.generateMd5String(length)
	case "GenerateTemplateHex":
		retVal = gn.generateTemplateHex(template)
	case "GenerateTemplate":
		retVal = gn.generateTemplate(template)
	case "GeneratePrevString":
		retVal = pb.previouslyGenerated[length]
		//fmt.Println(fmt.Sprintf("%d", length) + "<<<<<<<<<<")
	}
	//store
	pb.previouslyGenerated = append(pb.previouslyGenerated, retVal)
	return retVal
}

func GetFromToLineWithStrings(content string, from string, to string) string {
	contentParts := strings.Split(content, "\n")
	partialString := ""
	start := false
	for _, line := range contentParts {
		if strings.Contains(line, from) {
			start = true
		}
		if start == true {
			partialString += line + "\n"
		}
		if strings.Contains(line, to) {
			break
		}
	}
	return partialString
}

func (pb *postBody) shouldBeExcluded(inputName, stepExcept string) bool {
	if len(stepExcept) == 0 {
		return false
	}
	if strings.Contains(stepExcept, inputName) == false {
		return false
	}
	return true
}

func (pb *postBody) allInputsFromTo(step StepToGetPostVar) {
	if len(step.FromLineWith) > 1 {
		formOfInterest := GetFromToLineWithStrings(pb.content, step.FromLineWith, step.ToLineWith)
		//fmt.Println(len(pb.content), formOfInterest, "@@@@@@")
		for _, line := range strings.Split(formOfInterest, "\n") {
			regName := regexp.MustCompile(`name=(?:"|')?([^\s"]+)"?`)
			regValue := regexp.MustCompile(`value="?([^"]+)"?`)
			name := regName.FindStringSubmatch(line)
			value := regValue.FindStringSubmatch(line)

			if name != nil && value != nil &&
				pb.shouldBeExcluded(name[1], step.Except) == false {
				pb.Inputs.Add(name[1], html.UnescapeString(value[1]))
			}
		}
	}
}

func (pb *postBody) cookieName(step StepToGetPostVar) {
	if len(step.CookieName) > 1 {
		pb.Inputs.Add(step.PostKey, pb.cookieJar.cookieMap[step.CookieName])
	}
}

func (pb *postBody) lineWith(step StepToGetPostVar) {
	if len(step.LineWith) > 0 {
		for _, line := range strings.Split(pb.content, "\n") {
			//fmt.Println(step.LineWith, line, "<=========")
			if strings.Contains(line, step.LineWith) {
				lineParts := strings.Split(line, step.SplitString)
				ind, _ := strconv.Atoi(step.TakeIndex)
				pb.Inputs.Add(step.PostKey, lineParts[ind])
			}
		}
	}
}

func (pb *postBody) urlEncode(URIString string) string {
	return url.QueryEscape(URIString)
}

func (pb *postBody) urlDecode(URIString string) string {
	decodedValue, err := url.QueryUnescape(URIString)
	if err != nil {
		URIString = decodedValue
	}
	return URIString
}

func (pb *postBody) toString() string {
	/* retString := ""
	for key, value := range pb.Inputs {
		retString += fmt.Sprintf("%s=%s&", pb.urlEncode(key), pb.urlEncode(value))
	} */
	//TODO: remove repetition post parameters
	if len(pb.jsonPostData) > 3 {
		return pb.jsonPostData
	}
	return pb.Inputs.Encode()
}

func (pb *postBody) Reset(step StepToGetPostVar) {
	if len(step.Reset) > 1 {
		pb.init()
	}
}
