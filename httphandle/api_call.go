package httphandle

import (
	"gost/api"
	"gost/cache"
	"gost/config"
	"gost/filter"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
)

var apiInterface interface{}

func SetApiInterface(interf interface{}) {
	apiInterface = interf
}

func PerformApiCall(handlerName string, rw http.ResponseWriter, req *http.Request, route *config.Route) {
	// Prepare data vector for an api/endpoint call
	inputs := make([]reflect.Value, 1)

	// Create the variables containing request data
	vars := createApiVars(req, rw, route)
	if vars == nil {
		return
	}

	// Try giving the response directly from the cache if available
	if cache.Status == cache.STATUS_ON {
		if cachedData := cache.QueryByRequest(vars.RequestForm, route.Pattern); cachedData != nil {
			if req.Method == api.GET {
				GiveApiResponse(cachedData.StatusCode, cachedData.Data, rw, req, route.Pattern, cachedData.ContentType, cachedData.File)
				return
			} else { // Invalidate the cache if a modification, deletion or addition was made to this endpoint
				cachedData.Invalidate()
			}
		}
	}

	// Populate the data vector for the api call
	inputs[0] = reflect.ValueOf(vars)

	// Perform the call on the corresponding endpoint and function
	// This is done by using reflection techniques
	var respObjects []reflect.Value
	apiMethod := reflect.ValueOf(apiInterface).MethodByName(route.Handlers[req.Method])

	// Check for zero value
	if apiMethod != *new(reflect.Value) {
		respObjects = apiMethod.Call(inputs)
	} else {
		GiveApiStatus(http.StatusInternalServerError, rw, req, route.Pattern)

		log.Println("The endpoint method is either inexistent or incorrectly mapped. Please check the server configuration files!")

		return
	}

	if respObjects == nil {
		GiveApiStatus(http.StatusInternalServerError, rw, req, route.Pattern)
		return
	}

	// Extract the response from the endpoint into a concrete type
	resp := respObjects[0].Interface().(api.ApiResponse)

	// Give the response to the api client
	respond(vars, &resp, rw, req, route.Pattern)
}

func respond(vars *api.ApiVar, resp *api.ApiResponse, rw http.ResponseWriter, req *http.Request, endpoint string) {
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusInternalServerError
		GiveApiMessage(resp.StatusCode, http.StatusText(resp.StatusCode), rw, req, endpoint)
	} else if len(resp.ErrorMessage) > 0 {
		GiveApiMessage(resp.StatusCode, resp.ErrorMessage, rw, req, endpoint)
	} else {
		if len(resp.ContentType) == 0 {
			resp.ContentType = CONTENT_JSON
		}

		GiveApiResponse(resp.StatusCode, resp.Message, rw, req, endpoint, resp.ContentType, resp.File)

		// Try caching the data only if a GET request was made
		if req.Method == api.GET && cache.Status == cache.STATUS_ON {
			cacheResponse(vars, resp, endpoint)
		}
	}
}

func cacheResponse(vars *api.ApiVar, resp *api.ApiResponse, endpoint string) {
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) || len(resp.Message) == 0 {
		return
	}

	cacheEntity := &cache.Cache{
		Query:       cache.MapKey(vars.RequestForm, endpoint),
		Data:        resp.Message,
		StatusCode:  resp.StatusCode,
		ContentType: resp.ContentType,
		File:        resp.File,
	}

	cacheEntity.Cache()
}

func createApiVars(req *http.Request, rw http.ResponseWriter, route *config.Route) *api.ApiVar {
	err, statusCode := filter.CheckMethodAndParseContent(req)
	if err != nil {
		GiveApiMessage(statusCode, err.Error(), rw, req, route.Pattern)
		return nil
	}

	body, err := convertBodyToReadableFormat(req.Body)
	if err != nil {
		GiveApiMessage(http.StatusBadRequest, err.Error(), rw, req, route.Pattern)
		return nil
	}

	vars := &api.ApiVar{
		RequestHeader:        req.Header,
		RequestForm:          req.Form,
		RequestContentLength: req.ContentLength,
		RequestBody:          body,
	}

	return vars
}

func convertBodyToReadableFormat(data io.ReadCloser) ([]byte, error) {
	return ioutil.ReadAll(data)
}
