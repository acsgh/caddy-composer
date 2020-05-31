package module

import (
	"github.com/andybalholm/cascadia"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/http"
	"strings"
)

const GET = "get"
const AttributeUrlKey = "data-webc-url"
const AttributeMethodKey = "data-webc-method"
const AttributeNameKey = "data-webc-name"
const AttributeBodyKey = "data-webc-body"

type ComposeContext struct {
	webComposer  *WebComposer
	httpClient   *http.Client
	httpRequest  *http.Request
	httpResponse *caddyhttp.ResponseRecorder
	cache        *Cache
}

func (ctx *ComposeContext) compose(payload string) (*string, error) {
	doc, err := parseString(&payload)

	if err != nil {
		return nil, err
	}

	err = ctx.composeNode(doc, doc)

	if err != nil {
		return nil, err
	}

	return renderToString(doc)
}

func (ctx *ComposeContext) composeNode(doc *html.Node, node *html.Node) error {
	defaultMethod := GET

	divs := cascadia.MustCompile("div").MatchAll(node)
	for _, div := range divs {
		url := attr(div, AttributeUrlKey, nil)
		method := attr(div, AttributeMethodKey, &defaultMethod)
		name := attr(div, AttributeNameKey, nil)
		body := attr(div, AttributeBodyKey, nil)

		if url != nil && name != nil {
			ctx.logCompositionInfo("composition request", url, method, name)

			webComponent, err := ctx.getWebComponent(method, url, body, name)

			if err != nil {
				ctx.logCompositionError("composition error", url, method, name, err)
			} else {
				_ = ctx.replaceComponent(doc, webComponent, div)
			}
		}
	}

	return nil
}

func (ctx *ComposeContext) replaceComponent(doc *html.Node, component *WebComponent, dst *html.Node) error {
	err := ctx.composeNode(doc, component.content)

	if err != nil {
		return err
	}

	ctx.handoverResponseHeader(component.headers)
	replaceContent(dst, component.content)

	head := cascadia.MustCompile("head").MatchFirst(doc)
	attachIfRequired(head, "link", "href", component.stylesheets)

	body := cascadia.MustCompile("body").MatchFirst(doc)
	attachIfRequired(body, "script", "src", component.scripts)

	return nil
}

func (ctx *ComposeContext) handoverResponseHeader(header *http.Header) {
	for key, values := range *header {
		if ctx.mustHandoverHeader(key) {
			response := *ctx.httpResponse

			for _, value := range values {
				if !containsHeaderValue(response.Header(), key, value) {
					response.Header().Add(key, value)
				}
			}
		}
	}
}

func (ctx ComposeContext) getWebComponent(method *string, url *string, body *string, name *string) (*WebComponent, error) {
	source := newSource(method, url, body)

	loadedSource, _ := ctx.webComposer.cache.get(source.id)

	if loadedSource == nil {
		loadedSource, _ = ctx.cache.get(source.id)
	}

	if loadedSource == nil {
		err := source.load(ctx)

		if err != nil {
			return nil, err
		}

		loadedSource = source
		ctx.cache.set(loadedSource, nil)

		if loadedSource.cachedUntil != nil {
			ctx.webComposer.cache.set(loadedSource, loadedSource.cachedUntil)
		}

		if *loadedSource.responseStatusCode != 200 {
			err := errors.Errorf("The remote response was %d", *loadedSource.responseStatusCode)
			return nil, err
		}
	}

	if loadedSource == nil {
		return nil, errors.Errorf("Component source invalid")
	}

	return loadedSource.getWebComponent(name)
}

func (ctx *ComposeContext) mustHandoverHeader(key string) bool {
	return strings.HasPrefix(key, "X-") ||
		strings.EqualFold(key, "Set-Cookie") ||
		strings.EqualFold(key, "Cookie") ||
		strings.EqualFold(key, "Authorization")
}
