package common

import (
	"crypto/tls"
	"net"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/rs/xid"
)

func IsEmpty(s string) bool {
	s1 := strings.TrimSpace(s)
	return len(s1) == 0
}

func MakeHttpClient(timeout int) *http.Client {

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	var client = &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	return client
}

func getLastPath(s string, limit int) string {

	index := 0
	dir := s
	var arr []string

	for !IsEmpty(dir) {
		if index >= limit {
			break
		}
		index++
		arr = append([]string{path.Base(dir)}, arr...)
		dir = path.Dir(dir)
	}
	return path.Join(arr...)
}

func GetCallerInfo(offset int) (string, string, int) {

	pc := make([]uintptr, 15)
	n := runtime.Callers(offset, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	function := getLastPath(frame.Function, 1)
	file := getLastPath(frame.File, 3)
	line := frame.Line

	return function, file, line
}

func HasElem(s interface{}, elem interface{}) bool {

	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}
	return false
}

func GetGuid() string {
	guid := xid.New()
	return guid.String()
}
