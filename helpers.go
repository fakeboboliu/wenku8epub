package main

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
)

var dev = false

func noErr(err error) {
	if err != nil {
		if dev {
			log.Panic(err)
		} else {
			log.Fatal(err)
		}
	}
}

func check(ret interface{}, err error) interface{} {
	noErr(err)
	return ret
}

func randUUID() string {
	const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	uuid := []byte("")
	for i := 36; i > 0; i-- {
		switch i {
		case 27, 22, 17, 12:
			uuid = append(uuid, 45) // 45 is "-"
		default:
			uuid = append(uuid, chars[rand.Intn(36)])
		}
	}
	return string(uuid)
}

func httpGetWithRetry(url string, times int) (*http.Response, error) {
	var outErr error
	for re := 0; re < times; re++ {
		resp, err := http.Get(url)
		if err == nil {
			return resp, nil
		}
		outErr = err
	}
	return nil, outErr
}
