package utils

import "regexp"

func IsAlphanumeric(s string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9]+$")
	return re.MatchString(s)
}
