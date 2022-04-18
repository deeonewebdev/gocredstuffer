package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type ComboPurifier struct {
	comboList     *[]string
	UsernameRegex []string
	PasswordRegex []string
}

func (cp *ComboPurifier) fromComboList(cmb *[]string) *ComboPurifier {
	cp.comboList = cmb
	return cp
}

func (cp *ComboPurifier) validUserPass(user, pass string) bool {
	for _, reg := range cp.UsernameRegex {
		regi := regexp.MustCompile(reg)
		if regi.MatchString(user) == false {
			return false
		}
	}

	for _, reg := range cp.PasswordRegex {
		regi := regexp.MustCompile(reg)
		if regi.MatchString(pass) == false {
			return false
		}
	}

	return true
}

func (cp *ComboPurifier) purify() []string {
	resultList := []string{}
	purifierContent := ReadFile("purifier-" + os.Args[2])[0]
	if purifierContent == "" {
		purifierContent = "{usernameRegex:[], passwordRegex:[]}"
	}
	err := json.Unmarshal([]byte(purifierContent), &cp)
	if err != nil {
		//fmt.Println(configContent)
		//fmt.Println(err)
		panic("invalid purifier json file")
	}
	for _, line := range *cp.comboList {
		if strings.Contains(line, "::") == true {
			lineParts := strings.Split(line, "::")
			user := lineParts[0]
			pass := lineParts[1]
			userLength := 25
			wholeUserLength := 60
			passLength := 25
			username := strings.Split(user, "@")[0]
			if len(username) > userLength || len(pass) > passLength ||
				len(user) > wholeUserLength || cp.validUserPass(username, pass) == false ||
				username == pass {
				fmt.Printf("bad combo %s %s\r", line, strings.Repeat(" ", len(line)+20))
				continue
			}
			line = strings.Replace(line, "\n", "", -1)
			fmt.Printf("added combo %s to brute list %s\r", line, strings.Repeat(" ", len(line)+20))
			//still here? add combo to list
			resultList = append(resultList, line)
		}
	}
	return resultList
}
