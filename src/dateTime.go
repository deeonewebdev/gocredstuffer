package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type datetime struct {
}

func NewDateTime() *datetime {
	return &datetime{}
}

func (dt *datetime) interpreteDate(placement string, format string) string {
	currentTime := time.Now()
	dateParts := strings.Split(currentTime.Format("2006-01-02"), "-")
	year := dateParts[0]
	month := dateParts[1]
	day := dateParts[2]
	var retString string
	switch placement {
	case "ymd":
		retString = fmt.Sprintf(format, year, month, day)
	case "mdy":
		retString = fmt.Sprintf(format, month, day, year)
	case "myd":
		retString = fmt.Sprintf(format, month, year, day)
	case "dmy":
		retString = fmt.Sprintf(format, day, month, year)
	case "dym":
		retString = fmt.Sprintf(format, day, year, month)
	}
	return retString
}

func (dt *datetime) interpreteTime(placement string, format string) string {
	currentTime := time.Now()
	dateParts := strings.Split(currentTime.Format("15:04:05"), ":")
	hour := dateParts[0]
	minute := dateParts[1]
	second := dateParts[2]
	var retString string
	switch placement {
	case "hms":
		retString = fmt.Sprintf(format, hour, minute, second)
	case "msh":
		retString = fmt.Sprintf(format, minute, second, hour)
	case "mhs":
		retString = fmt.Sprintf(format, minute, hour, second)
	case "smh":
		retString = fmt.Sprintf(format, second, minute, hour)
	case "shm":
		retString = fmt.Sprintf(format, second, hour, minute)
	}
	return retString
}

func (dt *datetime) replaceIn(templateStr string) string {
	//use regex to find all date time placeholders
	regie := regexp.MustCompile(`\{(date|time)-([ymdhs]{3})\|([^}]+)\}`)
	matches := regie.FindAllStringSubmatch(templateStr, -1)
	var search, replace string
	if matches != nil {
		for _, matchitem := range matches {
			search = matchitem[0]
			switch matchitem[1] {
			case "date":
				replace = dt.interpreteDate(matchitem[2], matchitem[3])
			case "time":
				replace = dt.interpreteTime(matchitem[2], matchitem[3])
			}
			templateStr = strings.ReplaceAll(templateStr, search, replace)
		}
	}
	return templateStr
}
