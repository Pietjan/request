package response

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httputil"
)

func Builder(handler HandlerFn) Handler {
	return &builder{
		handler: handler,
	}
}

func (b *builder) clone() *builder {
	return &builder{
		handler:    b.handler,
		conditions: b.conditions,
	}
}

type builder struct {
	handler    HandlerFn
	conditions []Condition
}

// When implements ResponseHandler.
func (b *builder) When(condition Condition) Handler {
	dupe := b.clone()
	dupe.conditions = append(dupe.conditions, condition)
	return dupe

}

// Handle implements ResponseHandler.
func (b *builder) Handle(res *http.Response) error {
	if len(b.conditions) == 0 {
		return b.handler(res)
	}

	for _, condition := range b.conditions {
		if condition(res) {
			if err := b.handler(res); err != nil {
				return err
			}
		}
	}

	return nil
}

func JsonHandler(v any) Handler {
	return Builder(func(res *http.Response) error {
		return json.NewDecoder(res.Body).Decode(v)
	})
}

func XmlHandler(v any) Handler {
	return Builder(func(res *http.Response) error {
		return xml.NewDecoder(res.Body).Decode(v)
	})
}

func Status(code int) Condition {
	return func(res *http.Response) bool {
		return res.StatusCode == code
	}
}

func StatusRange(min, max int) Condition {
	return func(res *http.Response) bool {
		return res.StatusCode >= min && res.StatusCode <= max
	}
}

var (
	StatusOK                  = Status(http.StatusOK)
	StatusCreated             = Status(http.StatusCreated)
	StatusAccepted            = Status(http.StatusAccepted)
	StatusBadRequest          = Status(http.StatusBadRequest)
	StatusUnauthorized        = Status(http.StatusUnauthorized)
	StatusForbidden           = Status(http.StatusForbidden)
	StatusNotFound            = Status(http.StatusNotFound)
	StatusUnprocessableEntity = Status(http.StatusUnprocessableEntity)
	StatusInternalServerError = Status(http.StatusInternalServerError)
	StatusBadGateway          = Status(http.StatusBadGateway)
)

var (
	Status10x = StatusRange(http.StatusContinue, http.StatusEarlyHints)
	Status20x = StatusRange(http.StatusOK, http.StatusIMUsed)
	Status30x = StatusRange(http.StatusMultipleChoices, http.StatusPermanentRedirect)
	Status40x = StatusRange(http.StatusBadRequest, http.StatusUnavailableForLegalReasons)
	Status50x = StatusRange(http.StatusInternalServerError, http.StatusNetworkAuthenticationRequired)
)

func Dump(res *http.Response) error {
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		return err
	}

	fmt.Printf("response dump:\n%s\n", string(b))
	return nil
}
