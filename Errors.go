package jsonapi

import (
	"encoding/json"
	"fmt"
)

type ErrorsPayload struct {
	Errors []*JSONAPIError `json:"errors"`
}

type JSONAPIError struct {
	//id: a unique identifier for this particular occurrence of the problem.
	ID string `json:"id"`
	//status: the HTTP status code applicable to this problem, expressed as a string value. This SHOULD be provided.
	Status int `json:"status"`
	//code: an application-specific error code, expressed as a string value.
	Code string `json:"code,omitempty"`
	//title: a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title"`
	//detail: a human-readable explanation specific to this occurrence of the problem. Like title, this field’s value can be localized.
	Detail string `json:"detail"`
	/*
		source: an object containing references to the primary source of the error. It SHOULD include one of the following members or be omitted:

		    pointer: a JSON Pointer [RFC6901] to the value in the request document that caused the error [e.g. "/data" for a primary data object, or "/data/attributes/title" for a specific attribute]. This MUST point to a value in the request document that exists; if it doesn’t, the client SHOULD simply ignore the pointer.
		    parameter: a string indicating which URI query parameter caused the error.
		    header: a string indicating the name of a single request header which caused the error.
	*/
	Source map[string]interface{} `json:"source,omitempty"`
	//meta: a meta object containing non-standard meta-information about the error.
	Meta map[string]interface{} `json:"meta,omitempty"`
	//links: a links object that MAY contain the following members:
	//
	//    about: a link that leads to further details about this particular occurrence of the problem. When derefenced, this URI SHOULD return a human-readable description of the error.
	//    type: a link that identifies the type of error that this particular error is an instance of. This URI SHOULD be dereferencable to a human-readable explanation of the general error.
	Links map[string]interface{} `json:"links,omitempty"`
}

func (e *JSONAPIError) Error() string {
	return fmt.Sprintf("Error: %v %s %s\n", e.Status, e.Title, e.Detail)
}

func MarshalErrors(errs []*JSONAPIError) ([]byte, error) {
	return json.Marshal(ErrorsPayload{errs})
}
