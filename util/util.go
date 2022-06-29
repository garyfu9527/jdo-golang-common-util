package util

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	Log *logrus.Logger
)

var logInit bool = false
var LogLevel string = "info"

func initLog() {
	if logInit {
		return
	}

	//初始化logrus
	Log = logrus.New()
	switch LogLevel {
	case "debug":
		Log.Level = logrus.DebugLevel
	case "info":
		Log.Level = logrus.InfoLevel
	case "error":
		Log.Level = logrus.ErrorLevel
	default:
		Log.Level = logrus.DebugLevel
	}
	Log.Formatter = &logrus.JSONFormatter{
		DisableHTMLEscape: true,
	}
	Log.Out = os.Stdout
	logInit = true
}

func PrintLogRus(level, funcName string, args ...interface{}) {
	initLog()

	switch level {
	case "trace":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Trace(args)
	case "debug":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Debug(args)
	case "info":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Info(args)
	case "warn":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Warn(args)
	case "error":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Error(args)
	case "fatal":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Fatal(args)
	case "panic":
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Panic(args)
	default:
		Log.WithFields(logrus.Fields{
			"funcName": funcName,
		}).Debug(args)
	}
}

func GetUserIdFromJwt(jwt string) string {
	if len(jwt) == 0 {
		return ""
	}

	splites := strings.Split(jwt, " ")
	if len(splites) <= 1 {
		return splites[1]
	}

	type JdoUserInfo struct {
		OpenId string `json:"openId"`
	}

	userInfos := strings.Split(splites[1], ".")
	if len(userInfos) >= 2 {
		des, err := base64.StdEncoding.DecodeString(userInfos[1])
		if err != nil {
			PrintLogRus("error", "GetUserIdFromJwt DecodeString error:", userInfos[1])
			return ""
		}
		var jdoUserInfo JdoUserInfo
		err = json.Unmarshal(des, &jdoUserInfo)
		if err != nil {
			PrintLogRus("error", "GetUserIdFromJwt json Unmarshal error:", err)
			return ""
		}
		return jdoUserInfo.OpenId
	}

	return ""
}

func GetRemoteIP(ipForwarded string) string {
	splits := strings.Split(ipForwarded, ";")
	return splits[0]
}

func GetWeight(x []float64) int {
	length := len(x)
	sum := 0.0
	for i := 0; i < length; i++ {
		sum += x[i]
	}
	randVal := randFloats(0.0, sum)
	idx := 0
	for i := 0; i < length; i++ {
		if randVal <= x[i] {
			idx = i
			break
		}
		randVal -= x[i]
	}

	return idx
}

func randFloats(min, max float64) float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randVal := r.Float64()
	return min + randVal*(max-min)
}
