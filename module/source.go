package module

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WebSource struct {
	id                 *string
	method             *string
	url                *string
	body               *string
	responseStatusCode *int
	responseHeaders    *http.Header
	responseContent    *string
	cachedUntil        *time.Time
	loadTime           *time.Duration
}

type WebComponent struct {
	source      *WebSource
	name        *string
	headers     *http.Header
	content     *html.Node
	scripts     []*html.Node
	stylesheets []*html.Node
}

func newSource(method *string, url *string, body *string) *WebSource {
	result := new(WebSource)

	result.method = method
	result.url = url
	result.body = body
	result.id = result.calculateId()

	return result
}

func (s *WebSource) load(c ComposeContext) error {
	ti := time.Now()
	c.logCompositionDebug("composition fetching remote", s.url, s.method)
	requestBody := ""

	if s.body != nil {
		requestBody = *s.body
	}
	bodyReader := strings.NewReader(requestBody)

	request, err := http.NewRequest(*s.method, *s.url, bodyReader)

	if err != nil {
		return err
	}

	c.handoverRequestHeader(request.Header)

	response, err := c.httpClient.Do(request)

	if err != nil {
		return err
	}

	data, err := io.ReadAll(response.Body)

	if err != nil {
		return err
	}

	err = response.Body.Close()

	if err != nil {
		return err
	}

	dataString := string(data)

	s.responseStatusCode = &response.StatusCode
	s.responseHeaders = &response.Header
	s.responseContent = &dataString
	duration := ti.Sub(time.Now())
	s.loadTime = &duration
	s.cachedUntil = s.calculateCachedUntil()

	return nil
}

func (s *WebSource) calculateCachedUntil() *time.Time {
	prefix := "max-age="
	value := s.responseHeaders.Get("Cache-Control")
	if strings.HasPrefix(value, prefix) {
		number, err := strconv.Atoi(strings.ReplaceAll(value, prefix, ""))

		if err == nil {
			result := time.Now()
			result = result.Add(time.Duration(number) * time.Second)
			return &result
		}
	}
	return nil
}

func (s *WebSource) calculateId() *string {
	hasher := sha256.New()

	if s.method != nil {
		hasher.Write([]byte(*s.method))
	}

	if s.url != nil {
		hasher.Write([]byte("-"))
		hasher.Write([]byte(*s.url))
	}

	if s.body != nil {
		hasher.Write([]byte("-"))
		hasher.Write([]byte(*s.body))
	}

	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return &hash
}

func (s *WebSource) getWebComponent(name *string) (*WebComponent, error) {
	doc, err := parseString(s.responseContent)

	if err != nil {
		return nil, err
	}

	var componentNode *html.Node

	divs := cascadia.MustCompile("div").MatchAll(doc)
	for _, div := range divs {
		url := attr(div, AttributeUrlKey, nil)
		componentName := attr(div, AttributeNameKey, nil)

		if url == nil && componentName != nil && *componentName == *name {
			componentNode = div
		}

	}

	if componentNode == nil {
		return nil, errors.Errorf("Component %s not found", *name)
	}

	component := new(WebComponent)
	component.name = name
	component.source = s
	component.headers = s.responseHeaders
	component.content = componentNode

	head := cascadia.MustCompile("head").MatchFirst(doc)

	if head != nil {
		component.stylesheets = cascadia.MustCompile("link").MatchAll(head)
	}

	body := cascadia.MustCompile("body").MatchFirst(doc)

	if body != nil {
		component.scripts = cascadia.MustCompile("script").MatchAll(body)
	}

	return component, nil
}

func (ctx *ComposeContext) handoverRequestHeader(header http.Header) {
	for key, values := range ctx.httpRequest.Header {
		if ctx.mustHandoverHeader(key) {
			for _, value := range values {
				if !containsHeaderValue(header, key, value) {
					header.Add(key, value)
				}
			}
		}
	}
}

func containsHeaderValue(header http.Header, targetKey string, targetValue string) bool {
	for key, values := range header {
		for _, value := range values {
			if strings.EqualFold(targetKey, key) && targetValue == value {
				return true
			}
		}
	}
	return false
}
