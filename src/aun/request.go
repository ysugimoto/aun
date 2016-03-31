package aun

import (
	"strings"
)

type Request struct {
	RequestType int
	Method      string
	Path        string
	Version     string
	Headers     map[string]string
}

func NewRequest(message string, requestType int) *Request {
	list := strings.Split(message, "\r\n")

	// first line is request line
	parts := strings.Split(list[0], " ")

	// sencond-after lines are headers
	headers := make(map[string]string)
	for _, l := range list[1:] {
		if len(l) == 0 {
			continue
		}
		spl := strings.Split(l, ": ")
		headers[spl[0]] = spl[1]
	}

	return &Request{
		RequestType: requestType,
		Method:      parts[0],
		Path:        parts[1],
		Version:     parts[2],
		Headers:     headers,
	}
}

func (r *Request) HasHeader(key string) (ok bool) {
	_, ok = r.Headers[key]

	return
}

func (r *Request) GetHeader(key string) (header string) {
	header, _ = r.Headers[key]

	return
}
