package utils

import (
	"net/url"
	"strings"
)

func ClearUTMParams(intputUrl *url.URL) {
	newQuery := url.Values{}
	for key, value := range intputUrl.Query() {
		if !strings.HasPrefix(key, "utm_") {
			newQuery[key] = value
		}
	}

	intputUrl.RawQuery = newQuery.Encode()
}
