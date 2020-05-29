package module

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"net/http"
)

const GET = "get"
const AttributeUrlKey = "data-webc-url"
const AttributeMethodKey = "data-webc-method"
const AttributeNameKey = "data-webc-name"

type ComposeContext struct {
	webComposer   *WebComposer
	httpClient    *http.Client
	httpRequest   *http.Request
	responseCache map[string]*Response
}

type WebComponent struct {
	method          *string
	url             *string
	name            *string
	body            *string
	content         *string
	responseHeaders *http.Header
}

func (ctx *ComposeContext) compose(payload string) (*string, error) {
	defaultMethod := GET

	doc, err := parseString(&payload)
	if err != nil {
		return nil, err
	}

	divs := cascadia.MustCompile("div").MatchAll(doc)
	for _, div := range divs {
		url := attr(div, AttributeUrlKey, nil)
		method := attr(div, AttributeMethodKey, &defaultMethod)
		name := attr(div, AttributeNameKey, nil)

		if url != nil && name != nil {
			ctx.logCompositionInfo("composition request", url, method, name)

			remoteResponse, err := ctx.getRemoteContent(method, url, nil)

			if err != nil {
				ctx.logCompositionError("composition error", url, method, name, err)
			}

			if remoteResponse.statusCode != 200 {
				err := errors.Errorf("The remote response was %s", remoteResponse.statusCode)
				ctx.logCompositionError("composition error", url, method, name, err)
			}

			//println(remoteResponse.statusCode)
			//println(*remoteResponse.body)
			//content := *method + " " + *url + " " + *name
			replaceContentWithString(div, remoteResponse.body)
		}
	}

	return renderToString(doc)
}

func (c ComposeContext) getRemoteContent(method *string, url *string, body *string) (*Response, error) {
	requestHash := hash(method, url, body)
	cachedResponse := c.responseCache[requestHash]

	if cachedResponse != nil {
		return cachedResponse, nil
	} else {
		response, err := c.execRemoteContent(method, url, body)

		if err != nil {
			return nil, err
		}

		c.responseCache[requestHash] = response
		return response, nil
	}
}

func hash(method *string, url *string, body *string) string {
	input := ""

	hasher := sha256.New()

	if method != nil {
		input += *method
		hasher.Write([]byte(*method))
	}

	if url != nil {
		hasher.Write([]byte("-"))
		hasher.Write([]byte(*url))
	}

	if body != nil {
		hasher.Write([]byte("-"))
		hasher.Write([]byte(*body))
	}

	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
