package common

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	utils "github.com/devopsext/utils"
	"github.com/rs/xid"
	genUtils "github.com/uber/jaeger-client-go/utils"
)

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

	for !utils.IsEmpty(dir) {
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

func GetGuid() string {
	guid := xid.New()
	return guid.String()
}

func SpanIDHexToUint64(hex string) uint64 {

	iSpanID := new(big.Int)
	iSpanID, ok := iSpanID.SetString(hex, 16)
	if !ok {
		return 0
	}
	return iSpanID.Uint64()
}

func SpanIDUint64ToHex(n uint64) string {

	return fmt.Sprintf("%016x", n)
}

func SpanIDBytesToHex(bytes [8]byte) string {

	return hex.EncodeToString(bytes[:])
}

func TraceIDHexToUint64(hex string) uint64 {

	iTraceID := new(big.Int)
	iTraceID, ok := iTraceID.SetString(hex, 16)
	if !ok {
		return 0
	}
	return iTraceID.Uint64()
}

func TraceIDUint64ToHex(n uint64) string {

	return fmt.Sprintf("%032x", n)
}

func TraceIDBytesToHex(bytes [16]byte) string {

	return hex.EncodeToString(bytes[:])
}

var generator = genUtils.NewRand(time.Now().UnixNano())
var pool = sync.Pool{
	New: func() interface{} {
		return rand.NewSource(generator.Int63())
	},
}

func randomNumber() uint64 {
	generator := pool.Get().(rand.Source)
	number := uint64(generator.Int63())
	pool.Put(generator)
	return number
}

func NewTraceID() string {
	return TraceIDUint64ToHex(randomNumber())
}

func NewSpanID() string {
	return SpanIDUint64ToHex(randomNumber())
}

func GetKeyValues(s string) map[string]string {

	env := utils.GetEnvironment()
	pairs := strings.Split(s, ",")

	var m map[string]string = make(map[string]string)

	for _, p := range pairs {

		if utils.IsEmpty(p) {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		k := strings.TrimSpace(kv[0])
		if len(kv) > 1 {
			v := strings.TrimSpace(kv[1])
			if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
				ed := strings.SplitN(v[2:len(v)-1], ":", 2)
				e, d := ed[0], ed[1]
				v = env.Get(e, "").(string)
				if v == "" && d != "" {
					v = d
				}
			}
			m[k] = v
		} else {
			m[k] = ""
		}
	}
	return m
}

func MapToArray(m map[string]string) []string {

	var arr []string
	if m == nil {
		return arr
	}
	for k, v := range m {
		if utils.IsEmpty(v) {
			arr = append(arr, k)
		} else {
			arr = append(arr, fmt.Sprintf("%s=%v", k, v))
		}
	}
	return arr
}
