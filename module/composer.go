package module

import (
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"net/http"
)

const GET = "get"
const AttributeUrlKey = "data-webc-url"
const AttributeMethodKey = "data-webc-method"
const AttributeNameKey = "data-webc-name"
const AttributeBodyKey = "data-webc-body"

type ComposeContext struct {
	webComposer *WebComposer
	httpClient  *http.Client
	httpRequest *http.Request
	cache       *Cache
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
		body := attr(div, AttributeBodyKey, nil)

		if url != nil && name != nil {
			ctx.logCompositionInfo("composition request", url, method, name)

			webComponent, err := ctx.getWebComponent(method, url, body, name)

			if err != nil {
				ctx.logCompositionError("composition error", url, method, name, err)
			} else {
				_ = replaceComponent(doc, webComponent, div)
			}
		}
	}

	return renderToString(doc)
}

func replaceComponent(doc *html.Node, component *WebComponent, dst *html.Node) error {
	_ = replaceContent(dst, component.content)

	head := cascadia.MustCompile("head").MatchFirst(doc)

	if head != nil {
		for _, stylesheet := range component.stylesheets {
			_ = appendContent(head, stylesheet)
		}
	}

	body := cascadia.MustCompile("body").MatchFirst(doc)

	if body != nil {
		for _, script := range component.scripts {
			_ = appendContent(head, script)
		}
	}

	return nil
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
