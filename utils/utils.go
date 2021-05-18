// Copyright (c) 2018-present Papo. All Rights Reserved.
// See LICENSE.txt for license information.

package utils

import (
	//"bitbucket.org/enesyteam/papo-server/model"
	//"bitbucket.org/enesyteam/papo-server/utils/fileutils"
	"net"
	"net/http"
	"net/url"
	//"os"
	"regexp"
	"strings"
)

func StringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// RemoveStringFromSlice removes the first occurrence of a from slice.
func RemoveStringFromSlice(a string, slice []string) []string {
	for i, str := range slice {
		if str == a {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// RemoveStringsFromSlice removes all occurrences of strings from slice.
func RemoveStringsFromSlice(slice []string, strings ...string) []string {
	newSlice := []string{}

	for _, item := range slice {
		if !StringInSlice(item, strings) {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func StringArrayIntersection(arr1, arr2 []string) []string {
	arrMap := map[string]bool{}
	result := []string{}

	for _, value := range arr1 {
		arrMap[value] = true
	}

	for _, value := range arr2 {
		if arrMap[value] {
			result = append(result, value)
		}
	}

	return result
}

//func FileExistsInConfigFolder(filename string) bool {
//	if len(filename) == 0 {
//		return false
//	}
//
//	if _, err := os.Stat(fileutils.FindConfigFile(filename)); err == nil {
//		return true
//	}
//	return false
//}

func RemoveDuplicatesFromStringArray(arr []string) []string {
	result := make([]string, 0, len(arr))
	seen := make(map[string]bool)

	for _, item := range arr {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

func GetIpAddress(r *http.Request, trustedProxyIPHeader []string) string {
	address := ""

	for _, proxyHeader := range trustedProxyIPHeader {
		header := r.Header.Get(proxyHeader)
		if len(header) > 0 {
			addresses := strings.Fields(header)
			if len(addresses) > 0 {
				address = strings.TrimRight(addresses[0], ",")
			}
		}

		if len(address) > 0 {
			return address
		}
	}

	if len(address) == 0 {
		address, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return address
}

func GetHostnameFromSiteURL(siteURL string) string {
	u, err := url.Parse(siteURL)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// SanitizeName sanitizes user names in an email
func getNCharactersOfString(message string, limit int) string {
	// Đôi khi trong tin nhắn có chứa ký tự \n "" sẽ gây ra lỗi
	// vì thế cần loại bỏ các ký tự này ra khỏi kết quả
	var snippetBlacklist = regexp.MustCompile(`(&|>|<|\/|:|\n|\r)*`)
	snippetBlacklist.ReplaceAllString(message, "")
	result := message
	chars := 0
	for i := range message {
		if chars >= limit {
			result = message[:i] + "..."
			break
		}
		chars++
	}
	return result
}

func GetSnippet(s string) string {
	return getNCharactersOfString(s, 120)
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

