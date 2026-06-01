package api

import (
	"bytes"
	"decentraland-data-downloader-v4/packages/helpers"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func payloadToQuery(payload map[string]interface{}, prefix string) (queryParamsArr []string) {
	queryParamsArr = make([]string, 0)
	for key, value := range payload {
		vKey := key
		if prefix != "" {
			vKey = prefix + "." + key
		}
		if reflect.TypeOf(value).Kind() == reflect.Map {
			valueQParamsArr := payloadToQuery(value.(map[string]interface{}), vKey)
			queryParamsArr = append(queryParamsArr, valueQParamsArr...)
			continue
		}
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			for _, v := range value.([]string) {
				strV := fmt.Sprintf("%v", v)
				vv := url.QueryEscape(strV)
				queryParamsArr = append(queryParamsArr, vKey+"="+vv)
			}
			continue
		}
		vv := url.QueryEscape(fmt.Sprintf("%v", value))
		queryParamsArr = append(queryParamsArr, vKey+"="+vv)
	}
	return
}

func SendHttpRequest(resourceUrl, method string, headers map[string]string, payload map[string]any, output interface{}) (err error) {
	_url := resourceUrl
	var _body io.Reader
	if method == "GET" || method == "DELETE" {
		if payload != nil && len(payload) > 0 {
			queryParamsArr := payloadToQuery(payload, "")
			queryParamsStr := strings.Join(queryParamsArr, "&")
			_url = _url + "?" + queryParamsStr
		}
	} else {
		var jsonPayload []byte
		jsonPayload, err = json.Marshal(payload)
		if err != nil {
			return
		}
		_body = bytes.NewBuffer(jsonPayload)
	}
	var req *http.Request
	req, err = http.NewRequest(method, _url, _body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if headers != nil && len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var resBody []byte
	resBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		resJson := make(map[string]interface{})
		err = json.Unmarshal(resBody, &resJson)
		if err != nil {
			return
		}

		errorsObj, exists := resJson["errors"]
		if !exists {
			err = fmt.Errorf("request failed with status code %d", statusCode)
		} else {
			errorsArr := errorsObj.([]any)
			errorsArr2 := helpers.ArrayMap(errorsArr, func(t any) (bool, string) {
				return true, t.(string)
			}, true, "")
			errorMsg := strings.Join(errorsArr2, "|")
			if !strings.Contains(errorMsg, "not found") {
				return errors.New(errorMsg)
			} else {
				return nil
			}
		}
	}

	err = json.Unmarshal(resBody, output)
	return
}
