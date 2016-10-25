package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeXMLResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	response := Stock{}
	if err := xml.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func decodeJSONResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	response := Stock{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response, nil
}

func encodeRequest(_ context.Context, req *http.Request, _ interface{}) error {
	buf := bytes.NewBuffer([]byte{})
	req.Body = ioutil.NopCloser(buf)
	return nil
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeXMLResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	return xml.NewEncoder(w).Encode(response)
}
