package util

import (
	"bytes"
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"net/http"
	"xmediaEmu/pkg/log"
)

// Warn: AS port write directly here.
func PostAsCallbackUrl(url string, params map[string]interface{}) {
	status, resp, _ := Post(url, params)
	if status != 200 {
		log.Logger.Errorf("Post %v failed", url)
		return
	}
	log.Logger.Infof("Post %v OK, body:%v", url, resp)
}

func Post(url string, params map[string]interface{}) (int, *simplejson.Json, error) {
	bytesParams, err := json.Marshal(params)
	if err != nil {
		println(err)
		return 0, nil, err
	}
	return PostJson(url, bytesParams)
}

func PostJson(url string, bytesParams []byte) (int, *simplejson.Json, error) {
	request, err := http.NewRequest("POST", url, bytes.NewReader(bytesParams))
	if err != nil {
		println(err)
		return 0, nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		println(err)
		return 0, nil, err
	}

	data, err := simplejson.NewFromReader(resp.Body)
	return resp.StatusCode, data, err
}