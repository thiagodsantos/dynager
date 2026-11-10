package httpapi

import "strings"

func ParseOrigins(value string) []string {
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" && origin != "*" {
			origins = append(origins, origin)
		}
	}
	return origins
}
