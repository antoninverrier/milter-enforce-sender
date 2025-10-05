package main

import "strings"

func stringCaseCompare(s1 string, s2 string, cs bool) bool {

	if cs {
		return (s1 == s2)
	} else {
		return strings.EqualFold(s1, s2)
	}

}
