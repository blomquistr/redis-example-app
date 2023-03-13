package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
)

// a struct to hold our error messages and the corresponding http
// status code - this makes returning values from our decode function
// easier to handle on the caller's end. Wrapping your return values
// in a struct is generally a good practice for feeding a caller with
// known
type malformedRequest struct {
	status int
	msg    string
}

// a function for our malformedRequest errors to return the error
// message; much like a getter
func (mr *malformedRequest) Error() string {
	return mr.msg
}

// because many of our handlers are going to decode JSON, we'll want a
// handler to wrap that entire process for us. This will also let us take
// care of things like checking headers and error handling gracefully
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// First, we'll check the header of the request to make sure
	// it has the right content-type. We're using the gddo/httputil/header
	// library to perform this check, which will allow the check
	// to work even if the client includes bonus information or
	// an unexpected charset
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	// We'll use http.MaxBytesReader to enforce a maximum read size
	// from the response body. A request larger than that will now
	// cause an exception.
	r.Body = http.MaxBytesReader(w, r.Body, int64(config.getMaxBodySize()))

	// Setup the decoder and call DisallowUnknownFields() to cause Decode()
	// to return an unknown field error if it encounters unexpected extra
	// fields in the JSON body. Strictly speaking, it returns an error for
	// "keys which do not match any non-ignored, exported fields in the
	// desination"
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Here we'll handle any syntax errors and provide the end user
		// feedback that they can actually do something with. This is
		// easier for the caller than a generic error
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// our next case is if Decode() returns an io.ErrUnexpectedEOF
		// if the JSON is badly formatted or has incorrect JSON syntax.
		// There is an open issue regarding this at:
		// https://github.com/golang/go/issues/25956
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// next, check to see if we had an unmarshalling error for bad
		// type assignment; for example, if we're assigning a String to
		// an Int value in our Struct.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// Catch the error caused by unexpected fields in the request body
		// Extract the field anme from the error message and include it in
		// our message back to the user; this tells the user what to fix
		// and makes a better user experience
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// an io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		// Catch the error caused by a request body being too large; this
		// is kind of important, and tells the user to use a smaller message
		// which is particularly relevant for the limited size string of
		// Redis
		case err.Error() == "http: request body too large":
			msg := fmt.Sprintf("Request body must not be larger than %d", config.getMaxBodySize())
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		// Default response; log the error, send a 500
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

// TODO: Move the logic for rolling up a JSON object from the main server to this
// helper once I've figured out what that code looks like
func encodeJSONBody(w http.ResponseWriter, dst interface{}) error {
	// first, lets marshal our struct into a []byte; we don't want to set
	// the header yet, though, as our response will return an error, not
	// JSON, if the marshaling fails. We could also configure our API to
	// always return JSON, but that's a little bit of a different lesson
	resp, err := json.Marshal(dst)
	if err != nil {
		return err
	}

	// OK, we have successfully marshaled our response; we know we're
	// sending valid JSON back. Time to set the header, then write the
	// resposne back to the ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

	// no errors, so we can safely return nil
	return nil
}
