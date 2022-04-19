package main

import (
	"html"
	"strings"
)

type ComboOrProxyIsGood struct {
	cmb     *combo
	content string
	header  *Header
}

func (copis *ComboOrProxyIsGood) fromCombo(cmb *combo) *ComboOrProxyIsGood {
	copis.cmb = cmb
	return copis
}

func (copis *ComboOrProxyIsGood) withHeader(header *Header) *ComboOrProxyIsGood {
	copis.header = header
	return copis
}

func (copis *ComboOrProxyIsGood) fromContent(content string) *ComboOrProxyIsGood {
	copis.content = content
	return copis
}

func (copis *ComboOrProxyIsGood) isGood(criterii string) bool {
	if strings.Contains(criterii, "::") == false {
		return true
	}
	criterium := strings.Split(criterii, "^OR^")
	var isGood bool
criteriaLoop:
	for _, criteria := range criterium {
		criteriaParts := strings.Split(criteria, "::")
		separator := "|"
		/* if strings.Contains(criteriaParts[1], separator) == false {
			separator = "&"
		} */
		subCriteria := strings.Split(criteriaParts[1], separator)
		toSearchParts := strings.Split(criteriaParts[0], "!")
		toSearch := toSearchParts[0]
		compareTo := true
		if len(toSearchParts) == 2 {
			compareTo = false
			toSearch = toSearchParts[1]
		}

		for _, criterium := range subCriteria {
			switch toSearch {
			case "cookie":
				//check cookies for string match
				isGood = strings.Contains(
					URLDecodeString(
						copis.header.CookieJar.ToString()),
					criterium) == compareTo

			case "location":
				//check locations archive for string match
				isGood = strings.Contains(
					URLDecodeString(
						strings.Join(
							copis.header.Location.Archive, "\n")),
					criterium) == compareTo
			case "header":
				isGood = strings.Contains(copis.header.responseToString(), criterium) == compareTo
			default:
				//check content for strings match
				cont := URLDecodeString(
					html.UnescapeString(copis.content))

				isGood = strings.Contains(
					cont,
					criterium) == compareTo
			}

			if isGood == compareTo {
				break criteriaLoop
			}
		}
	}

	return isGood
}
