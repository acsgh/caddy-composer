package module

import (
	"bytes"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func init() {
	caddy.RegisterModule(WebComposer{})
	httpcaddyfile.RegisterHandlerDirective("web-composer", parseCaddyfile)
}

// WebComposer is an example; put your own type here.
type WebComposer struct {
	logger    *zap.Logger
	cache     *Cache
	MIMETypes []string `json:"mime_types,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (WebComposer) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.web-composer",
		New: func() caddy.Module { return new(WebComposer) },
	}
}

// Provision implements caddy.Provisioner.
func (w *WebComposer) Provision(ctx caddy.Context) error {
	w.logger = ctx.Logger(w) // g.logger is a *zap.Logger
	w.logger.Info("Starting Web-Composer module")

	if w.MIMETypes == nil {
		w.MIMETypes = defaultMIMETypes
	}

	w.cache = w.createCache()

	return nil
}

// Validate implements caddy.Validator.
func (w *WebComposer) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (w WebComposer) ServeHTTP(rw http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	// shouldBuf determines whether to execute templates on this response,
	// since generally we will not want to execute for images or CSS, etc.
	shouldBuf := func(status int, header http.Header) bool {
		ct := header.Get("Content-Type")
		for _, mt := range w.MIMETypes {
			if strings.Contains(ct, mt) {
				return true
			}
		}
		return false
	}
	rec := caddyhttp.NewResponseRecorder(rw, buf, shouldBuf)

	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	if !rec.Buffered() {
		return nil
	}

	err = w.composeRequest(rec, r)
	if err != nil {
		return err
	}

	rec.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	rec.Header().Del("Accept-Ranges") // we don't know ranges for dynamically-created content
	rec.Header().Del("Last-Modified") // useless for dynamic content since it's always changing

	// we don't know a way to quickly generate etag for dynamic content,
	// and weak etags still cause browsers to rely on it even after a
	// refresh, so disable them until we find a better way to do this
	rec.Header().Del("Etag")

	return rec.WriteResponse()
}

func (w *WebComposer) composeRequest(rr caddyhttp.ResponseRecorder, r *http.Request) error {
	buffer := rr.Buffer()

	composeContext := w.createContext(r, &rr)

	result, err := composeContext.compose(buffer.String())

	if err != nil {
		return err
	}

	buffer.Reset()
	_, err = buffer.Write([]byte(*result))

	return err
}

func (w *WebComposer) createContext(request *http.Request, response *caddyhttp.ResponseRecorder) *ComposeContext {
	composeContext := new(ComposeContext)
	composeContext.webComposer = w
	composeContext.httpClient = w.newHttpClient()
	composeContext.cache = w.createCache()
	composeContext.httpRequest = request
	composeContext.httpResponse = response
	return composeContext
}

func (w WebComposer) newHttpClient() *http.Client {
	result := new(http.Client)
	return result
}

// parseCaddyfile unmarshals tokens from h into a new WebComposer.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m WebComposer
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (w *WebComposer) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*WebComposer)(nil)
	_ caddy.Validator             = (*WebComposer)(nil)
	_ caddyhttp.MiddlewareHandler = (*WebComposer)(nil)
	_ caddyfile.Unmarshaler       = (*WebComposer)(nil)
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var defaultMIMETypes = []string{
	"text/html",
	"text/plain",
	"text/markdown",
}
