package examples

import (
	"github.com/pietjan/request"
	"github.com/pietjan/request/response"
)

func BasicJsonRequest() {
	var users []map[string]any
	request.Get(`https://api.github.com/users`).
		JsonResponse(&users)
}

func ParameterJsonRequest() {
	var users []map[string]any
	request.Get(`https://api.github.com/users`).
		AddParmeter(`page_size`, `5`).
		JsonResponse(&users)
}

func ConditonalJsonRequest() {
	var users []map[string]any
	var clientError map[string]any
	var serverError map[string]any

	request.Get(`https://api.github.com/users`).
		Handle(
			response.JsonHandler(&users).When(response.Status20x),
			response.JsonHandler(&clientError).When(response.Status40x),
			response.JsonHandler(&serverError).When(response.Status50x),
		)
}
