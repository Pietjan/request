package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	MimeTypeJson           = `application/json`
	MimeTypeXml            = `application/xml`
	MimeTypeFormUrlEncoded = `application/x-www-form-urlencoded`
)

const (
	HeaderAccept      = `Accept`
	HeaderContentType = `Content-Type`
)

type RequestBuilder interface {
	Get(url string) GetBuilder
	Post(url string) PostBuilder
	Patch(url string) PatchBuilder
	Put(url string) PutBuilder
	Delete(url string) DeleteBuilder
}

type NoPayloadBuilder interface {
	CommonBuilder
}

type PayloadBuilder interface {
	CommonBuilder
	JsonBody(v any) CommonBuilder
	XmlBody(v any) CommonBuilder
	Body(r io.Reader) CommonBuilder
}

type GetBuilder interface {
	NoPayloadBuilder
}

type PostBuilder interface {
	PayloadBuilder
	FormBody(form Form) PostBuilder
}

type PatchBuilder interface {
	PayloadBuilder
}

type PutBuilder interface {
	PayloadBuilder
}

type DeleteBuilder interface {
	NoPayloadBuilder
}

type CommonBuilder interface {
	Context(ctx context.Context) CommonBuilder
	AddParmeter(key string, value string) CommonBuilder
	SetParmeter(key string, value string) CommonBuilder
	AddHeader(key string, value string) CommonBuilder
	SetHeader(key string, value string) CommonBuilder
	Build() (*http.Request, error)
	Do() (*http.Response, error)
	Handle(...ResponseHandler) error
	JsonResponse(v any) error
	XmlResponse(v any) error
	Before(fn BeforeFn) CommonBuilder
	After(fn AfterFn) CommonBuilder
}

type BeforeFn = func(req *http.Request) error
type AfterFn = func(res *http.Response) error

type ResponseHandler interface {
	Handle(res *http.Response) error
}

type FormBuilder interface {
	Add(k, v string) FormBuilder
	Set(k, v string) FormBuilder
}

type Form struct {
	Values url.Values
}

func Dump(req *http.Request) error {
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
	}

	fmt.Printf("request dump:\n%s\n", string(b))

	return nil
}
