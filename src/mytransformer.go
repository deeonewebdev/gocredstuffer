package main

import (
	"fmt"
	"strings"
)

type myTransformer struct {
	Template        string
	Input           string
	Email           string
	Username        string
	UsernameLetters string
	Domain          string
	DomainLetters   string
	Provider        string
	ProviderLetters string
	Tld             string
}

func (tf *myTransformer) getLetters(input string) string {
	letters := ""
	alphabets := "abcdefghijklmnopqrstuvwxyz"
	inputAsList := strings.Split(input, "")
	for _, char := range inputAsList {
		if strings.Contains(alphabets, char) {
			letters += char
		}
	}
	return letters
}

func (tf *myTransformer) usingInput(input string) *myTransformer {
	tf.Input = strings.ToLower(input)
	if strings.Contains(tf.Input, "@") {
		inputParts := strings.Split(tf.Input, "@")
		tf.Username = strings.ToLower(inputParts[0])
		tf.Domain = strings.ToLower(inputParts[1])
		if strings.Contains(tf.Domain, ".") {
			domainParts := strings.Split(tf.Domain, ".")
			tf.Provider = strings.ToLower(domainParts[0])
			tf.Tld = strings.ToLower(domainParts[1])
		}
		tf.UsernameLetters = tf.getLetters(tf.Username)
		tf.DomainLetters = tf.getLetters(tf.Domain)
		tf.ProviderLetters = tf.getLetters(tf.Provider)
	}
	return tf
}

func (tf *myTransformer) replace(variable, value string) {
	tf.Template = strings.Replace(tf.Template, variable, value, -1)
}

func (tf *myTransformer) reverse(s string) string {
	var result string
	for _, v := range s {
		result = string(v) + result
	}
	return result
}

func (tf *myTransformer) transformItem(templateVars []string, value string) {
	if strings.Contains(tf.Template, "%") {
		tf.replace(templateVars[0], value)
		tf.replace(templateVars[1], strings.ToTitle(value))
		tf.replace(templateVars[2], tf.reverse(value))
		tf.replace(templateVars[3], fmt.Sprintf("%s%s%s", strings.ToUpper(string(value[0:1])),
			strings.ToLower(string(value[1:len(value)-1])),
			strings.ToUpper(value[len(value)-1:])))
		tf.replace(templateVars[4], tf.reverse(strings.ToTitle(tf.reverse(value))))
		tf.replace(templateVars[5], strings.ToUpper(value))
	}
}

func (tf *myTransformer) transform(template string) string {
	tf.Template = template
	//tf.transformEmail()
	emailTransformList := []string{"%email%", "%Email%", "%reverse_email%", "%EmaiL%", "%emaiL%", "%EMAIL%"}
	userTransformList := []string{"%user%", "%User%", "%reverse_user%", "%UseR%", "%useR%", "%USER%"}
	domainTransformList := []string{"%domain%", "%Domain%", "%reverse_domain%", "%DomaiN%", "%domaiN%", "%DOMAIN%"}
	tldTransformList := []string{"%tld%", "%Tld%", "%reverse_tld%", "%TlD%", "%tlD%", "%TLD%"}
	providerTransformList := []string{"%provider%", "%Provider%", "%reverse_provider%", "%ProvideR%", "%provideR%", "%PROVIDER%"}
	userLettersTransformList := []string{"%userletters%", "%Userletters%", "%reverse_userletters%", "%UserletterS%",
		"%userletterS%", "%USERLETTERS%"}
	domainLettersTransformList := []string{"%domainletters%", "%Domainletters%", "%reverse_domainletters%", "%DomainletterS%",
		"%domainletterS%", "%DOMAINLETTERS%"}
	providerLettersTransformList := []string{"%providerletters%", "%Providerletters%", "%reverse_providerletters%",
		"%ProviderletterS%", "%providerletterS%", "%PROVIDERLETTERS%"}

	tf.transformItem(emailTransformList, tf.Input)
	tf.transformItem(userTransformList, tf.Username)
	tf.transformItem(userLettersTransformList, tf.UsernameLetters)
	tf.transformItem(domainTransformList, tf.Domain)
	tf.transformItem(domainLettersTransformList, tf.DomainLetters)
	tf.transformItem(providerTransformList, tf.Provider)
	tf.transformItem(providerLettersTransformList, tf.ProviderLetters)
	tf.transformItem(tldTransformList, tf.Tld)
	return tf.Template
}
