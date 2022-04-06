package common

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"path"
	"runtime"
	"sync"
	"time"

	utils "github.com/devopsext/utils"
	"github.com/rs/xid"
	genUtils "github.com/uber/jaeger-client-go/utils"
)

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
