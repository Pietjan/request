package request

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"github.com/pietjan/request/response"
)

func Builder(options ...func(*builder)) RequestBuilder {
	b := &builder{
		client:  http.DefaultClient,
		context: context.Background(),
		method:  http.MethodGet,
		url:     ``,
		header: Header{
			values: map[string][]string{},
		},
		parameter: Parameter{
			values: map[string][]string{},
		},
	}

	for _, fn := range options {
		fn(b)
	}

	return b
}

func Get(url string, options ...func(*builder)) GetBuilder {
	return Builder(options...).Get(url)
}

func Put(url string, options ...func(*builder)) PutBuilder {
	return Builder(options...).Put(url)
}

func Post(url string, options ...func(*builder)) PostBuilder {
	return Builder(options...).Post(url)
}

func Patch(url string, options ...func(*builder)) PatchBuilder {
	return Builder(options...).Patch(url)
}

func Delete(url string, options ...func(*builder)) DeleteBuilder {
	return Builder(options...).Delete(url)
}

func (b *builder) clone() *builder {
	return &builder{
		context:   b.context,
		client:    b.client,
		method:    b.method,
		url:       b.url,
		header:    b.header,
		parameter: b.parameter,
		body:      b.body,
		before:    b.before,
		after:     b.after,
		handlers:  b.handlers,
	}
}

func (b *builder) setContentTypeHeader(s string) {
	if len(b.header.values.Get(HeaderContentType)) > 0 {
		return
	}

	b.header.values.Set(HeaderContentType, s)
}

func (b *builder) setAcceptHeader(s string) {
	if len(b.header.values.Get(HeaderAccept)) > 0 {
		return
	}

	b.header.values.Set(HeaderAccept, s)
}

type Parameter struct {
	values url.Values
}

type Header struct {
	values http.Header
}

type builder struct {
	context   context.Context
	client    *http.Client
	method    string
	url       string
	header    Header
	parameter Parameter
	body      io.Reader
	before    []BeforeFn
	after     []AfterFn
	handlers  []ResponseHandler
}

func (b *builder) After(after AfterFn) CommonBuilder {
	dupe := b.clone()
	dupe.after = append(dupe.after, after)
	return dupe
}

func (b *builder) Before(before BeforeFn) CommonBuilder {
	dupe := b.clone()
	dupe.before = append(dupe.before, before)
	return dupe
}

func (b *builder) JsonHandler(v any) error {
	b.setAcceptHeader(MimeTypeJson)
	b.handlers = append(b.handlers, response.JsonHandler(v))
	_, err := b.Do()
	return err
}

func (b *builder) XmlHandler(v any) error {
	b.setAcceptHeader(MimeTypeXml)
	b.handlers = append(b.handlers, response.XmlHandler(v))
	_, err := b.Do()
	return err
}

func (b *builder) Handle(handlers ...ResponseHandler) error {
	b.handlers = handlers
	_, err := b.Do()
	return err
}

func (b *builder) AddHeader(key string, value string) CommonBuilder {
	dupe := b.clone()
	dupe.header.values.Add(key, value)
	return dupe
}

func (b *builder) AddParmeter(key string, value string) CommonBuilder {
	dupe := b.clone()
	dupe.parameter.values.Add(key, value)
	return dupe
}

func (b *builder) SetHeader(key string, value string) CommonBuilder {
	dupe := b.clone()
	dupe.header.values.Set(key, value)
	return dupe
}

func (b *builder) SetParmeter(key string, value string) CommonBuilder {
	dupe := b.clone()
	dupe.parameter.values.Set(key, value)
	return dupe
}

func (b *builder) Get(url string) GetBuilder {
	dupe := b.clone()
	dupe.url = url
	return dupe
}

func (b *builder) JsonBody(v any) CommonBuilder {
	dupe := b.clone()
	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		panic(err)
	}
	b.body = buf

	dupe.setContentTypeHeader(MimeTypeJson)

	return dupe
}

func (b *builder) XmlBody(v any) CommonBuilder {
	dupe := b.clone()
	buf := new(bytes.Buffer)

	if err := xml.NewEncoder(buf).Encode(v); err != nil {
		panic(err)
	}
	b.body = buf

	dupe.setContentTypeHeader(MimeTypeXml)

	return dupe
}

func (b *builder) FormBody(form Form) PostBuilder {
	dupe := b.clone()
	dupe.body = bytes.NewBufferString(form.Values.Encode())
	dupe.setContentTypeHeader(MimeTypeFormUrlEncoded)
	return dupe
}

func (b *builder) Body(r io.Reader) CommonBuilder {
	dupe := b.clone()
	dupe.body = r
	return dupe
}

func (b *builder) Patch(url string) PatchBuilder {
	dupe := b.clone()
	dupe.method = http.MethodPatch
	dupe.url = url
	return dupe
}

func (b *builder) Post(url string) PostBuilder {
	dupe := b.clone()
	dupe.method = http.MethodPost
	dupe.url = url
	return dupe
}

func (b *builder) Put(url string) PutBuilder {
	dupe := b.clone()
	dupe.method = http.MethodPut
	dupe.url = url
	return dupe
}

func (b *builder) Context(ctx context.Context) CommonBuilder {
	dupe := b.clone()
	dupe.context = ctx
	return dupe
}

func (b *builder) Build() (*http.Request, error) {
	url, err := url.Parse(b.url)
	if err != nil {
		return nil, err
	}

	url.RawQuery = b.parameter.values.Encode()

	req, err := http.NewRequestWithContext(b.context, b.method, url.String(), b.body)
	if err != nil {
		return nil, err
	}

	req.Header = b.header.values

	return req, nil
}

func (b *builder) Delete(url string) DeleteBuilder {
	dupe := b.clone()
	dupe.method = http.MethodDelete
	dupe.url = url
	return dupe
}

func (b *builder) Do() (*http.Response, error) {
	request, err := b.Build()
	if err != nil {
		return nil, err
	}

	for _, fn := range b.before {
		if err := fn(request); err != nil {
			return nil, err
		}
	}

	response, err := b.client.Do(request)
	if err != nil || len(b.handlers) == 0 {
		return response, err
	}

	for _, fn := range b.after {
		if err := fn(response); err != nil {
			return nil, err
		}
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	for _, handler := range b.handlers {
		res := *response
		if err := handler.Handle(&res); err != nil {
			return response, err
		}

		response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return response, err
}

func Client(client *http.Client) func(*builder) {
	return func(b *builder) {
		if client == nil {
			panic(`nil client`)
		}

		b.client = client
	}
}

func Before(fn BeforeFn) func(*builder) {
	return func(b *builder) {
		b.before = append(b.before, fn)
	}
}

func After(fn AfterFn) func(*builder) {
	return func(b *builder) {
		b.after = append(b.after, fn)
	}
}
