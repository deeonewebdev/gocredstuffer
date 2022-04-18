package main

import (
	"strings"
)

type UserAgent struct {
	Agent string
}

func NewUserAgent(agent string) *UserAgent {
	return &UserAgent{Agent: agent}
}

func (ua *UserAgent) isMobilePlatform() bool {
	return strings.Contains(strings.ToLower(ua.Agent), "mobile") ||
		strings.Contains(strings.ToLower(ua.Agent), "iphone") ||
		strings.Contains(strings.ToLower(ua.Agent), "android")
}

func (ua *UserAgent) isDesktopPlatform() bool {
	return ua.isMobilePlatform() == false &&
		(strings.Contains(strings.ToLower(ua.Agent), "linux") ||
			strings.Contains(strings.ToLower(ua.Agent), "windows") ||
			strings.Contains(strings.ToLower(ua.Agent), "darwin") ||
			strings.Contains(strings.ToLower(ua.Agent), "mac"))
}

func (ua *UserAgent) isUnknownPlatform() bool {
	return ua.isBot() == false
}

func (ua *UserAgent) isBot() bool {
	return ua.isDesktopPlatform() == false && len(ua.Agent) < 20
}

func (ua *UserAgent) getOperatingSystem() string {
	var retVal string
	switch {
	case strings.Contains(strings.ToLower(ua.Agent), "iphone"):
		retVal = "iphone"
	case strings.Contains(strings.ToLower(ua.Agent), "android"):
		retVal = "android"
	case strings.Contains(strings.ToLower(ua.Agent), "windows"):
		retVal = "windows"
	case strings.Contains(strings.ToLower(ua.Agent), "linux"):
		retVal = "linux"
	case strings.Contains(strings.ToLower(ua.Agent), "darwin"):
		retVal = "darwin"
	case strings.Contains(strings.ToLower(ua.Agent), "mac"):
		retVal = "mac"
	default:
		retVal = "unknown"
	}
	return retVal
}
