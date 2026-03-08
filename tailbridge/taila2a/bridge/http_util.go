package main

import "net/http"

func copyHeaders(dst http.Header, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
