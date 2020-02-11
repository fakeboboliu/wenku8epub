package helpers

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var dev = false

func NoErr(err error) {
	if err != nil {
		if dev {
			log.Panic(err)
		} else {
			log.Fatal(err)
		}
	}
}

func Check(ret interface{}, err error) interface{} {
	NoErr(err)
	return ret
}

func RandUUID() string {
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

var client *http.Client

func SetupProxy(proxyUrl string) {
	var tr *http.Transport
	if proxyUrl != "" {
		log.WithField("url", proxyUrl).Info("Loaded proxy")
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			log.Panic(err)
		}
		tr = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	} else {
		tr = &http.Transport{}
	}

	err := http2.ConfigureTransport(tr)
	if err != nil {
		panic(err)
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}
}

func HttpGetWithRetry(url string, times int) (*http.Response, error) {
	var outErr error
	for re := 0; re < times; re++ {
		resp, err := client.Get(url)
		if err == nil {
			return resp, nil
		}
		outErr = err
	}
	return nil, outErr
}
