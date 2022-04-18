package main

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
)

type Generator struct {
	seededRand *rand.Rand
	letters    string
	digits     string
	hexLetters string
	vowels     string
	consonants string
}

func (gn *Generator) init() *Generator {
	gn.seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	gn.letters = "abcdefghijklmnopqrstuvwxyz"
	gn.vowels = "aeiou"
	gn.consonants = "bcdfghjklmnpqrstvwxyz"
	gn.digits = "0123456789"
	gn.hexLetters = "abcdef"
	return gn
}

func (gn *Generator) stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[gn.seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (gn *Generator) generateTemplateHex(template string) string {
	var retString string
	for _, char := range template {
		charString := string(char)
		switch {
		case strings.Contains(gn.hexLetters, charString):
			//lowercase vowels
			retString += gn.stringWithCharset(1, gn.hexLetters)
		case strings.Contains(strings.ToUpper(gn.hexLetters), charString):
			//uppercase vowels
			retString += gn.stringWithCharset(1, strings.ToUpper(gn.hexLetters))
		case strings.Contains(gn.digits, charString):
			//lowercase digits
			retString += gn.stringWithCharset(1, gn.digits)
		default:
			retString += charString
		}
	}
	return retString
}

func (gn *Generator) generateTemplate(template string) string {
	var retString string
	for _, char := range template {
		charString := string(char)
		switch {
		case strings.Contains(gn.vowels, charString):
			//lowercase vowels
			retString += gn.stringWithCharset(1, gn.vowels)
		case strings.Contains(strings.ToUpper(gn.vowels), charString):
			//uppercase vowels
			retString += gn.stringWithCharset(1, strings.ToUpper(gn.vowels))
		case strings.Contains(gn.consonants, charString):
			//lowercase consonants
			retString += gn.stringWithCharset(1, gn.consonants)
		case strings.Contains(strings.ToUpper(gn.consonants), charString):
			//uppercase consonants
			retString += gn.stringWithCharset(1, strings.ToUpper(gn.consonants))
		case strings.Contains(gn.digits, charString):
			//lowercase digits
			retString += gn.stringWithCharset(1, gn.digits)
		default:
			retString += charString
		}
	}
	return retString
}

func (gn *Generator) generateLowString(length int) string {
	return gn.stringWithCharset(length, gn.letters)
}

func (gn *Generator) generateCapLowString(length int) string {
	return gn.stringWithCharset(length, gn.letters+strings.ToUpper(gn.letters))
}

func (gn *Generator) generateCapString(length int) string {
	return gn.stringWithCharset(length, strings.ToUpper(gn.letters))
}

func (gn *Generator) generateNumLowString(length int) string {
	return gn.stringWithCharset(length, gn.letters+gn.digits)
}

func (gn *Generator) generateNumCapLowString(length int) string {
	return gn.stringWithCharset(length, strings.ToUpper(gn.letters)+gn.letters+gn.digits)
}

func (gn *Generator) generateNumCapString(length int) string {
	return gn.stringWithCharset(length, strings.ToUpper(gn.letters)+gn.digits)
}

func (gn *Generator) generateNumString(length int) string {
	return gn.stringWithCharset(length, gn.digits)
}

func (gn *Generator) getMd5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (gn *Generator) generateMd5String(length int) string {
	return gn.getMd5Hash(gn.generateNumCapLowString(length))
}
