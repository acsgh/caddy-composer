package module

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type Response struct {
	statusCode int
	header     http.Header
	body       *string
}

func (ctx ComposeContext) execRemoteContent(method *string, url *string, body *string) (*Response, error) {
	requestBody := ""
	if body != nil {
		requestBody = *body
	}
	bodyReader := strings.NewReader(requestBody)

	request, err := http.NewRequest(*method, *url, bodyReader)

	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	response.Body.Close()

	dataString := string(data)

	result := new(Response)
	result.statusCode = response.StatusCode
	result.header = response.Header
	result.body = &dataString

	return result, nil
}
