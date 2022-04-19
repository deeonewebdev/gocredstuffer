package main

import (
	"net/url"
	"strings"
)

type Logger struct {
	cmb     *combo
	content *string
	header  *Header
}

func (log *Logger) withHeader(header *Header) *Logger {
	log.header = header
	return log
}

func (log *Logger) fromCombo(cmb *combo) *Logger {
	log.cmb = cmb
	return log
}

func (log *Logger) fromContent(content *string) *Logger {
	log.content = content
	return log
}

func (log *Logger) getHeader() *Header {
	switch {
	case log.cmb.wp != nil:
		return log.cmb.wp.getHeader()
	default:
		return log.cmb.nwp.getHeader()
	}
}

func (log *Logger) saveLog(add string) {
	filename := url.QueryEscape(log.cmb.username)
	var data string
	//add headers
	data += strings.Join(log.header.Response.HeadersArchive, "\n") + "\n"
	data += strings.Join(log.header.Request.HeadersArchive, "\n") + "\n"
	data += strings.Join(log.header.Location.Archive, "\n") + "\n"
	data += strings.Join(log.header.Location.RequestChain, "\n") + "\n"

	data += log.header.CookieJar.ToString()
	data += add + "\n\n"
	data += "\n\n\n\n\n"
	//data += *log.content
	writeToFile(filename+".txt", data)
}
