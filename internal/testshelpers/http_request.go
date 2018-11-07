// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package testshelpers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"lab.wtfteam.pro/wtfteam/lbtds/context"
)

// HTTPTestRequest executes HTTP request while testing and returns reply's
// body as bytes and HTTP code.
func HTTPTestRequest(t *testing.T, c *context.Context, reqBody []byte, reqHeaders map[string]string, method string, version string, path string, handler func(w http.ResponseWriter, r *http.Request)) ([]byte, int) {
	req := httptest.NewRequest(
		method,
		"http://"+c.Config.API.Address+":"+c.Config.API.Port+"/api/"+version+"/"+path+"/",
		bytes.NewBuffer(reqBody),
	)
	req.Header.Set("Content-Type", "application/json")
	for headerName, headerData := range reqHeaders {
		req.Header.Set(headerName, headerData)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	body, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err, "body read error")
	return body, rec.Code
}

// HTTPClearTestRequest is like HTTPTestRequest, but without context
func HTTPClearTestRequest(t *testing.T, address string, domain string, reqBody []byte, reqHeaders map[string]string, method string, handler func(w http.ResponseWriter, r *http.Request)) ([]byte, int) {
	req := httptest.NewRequest(
		method,
		address,
		bytes.NewBuffer(reqBody),
	)
	req.Host = domain
	for headerName, headerData := range reqHeaders {
		req.Header.Set(headerName, headerData)
	}
	rec := httptest.NewRecorder()
	handler(rec, req)
	body, err := ioutil.ReadAll(rec.Body)
	assert.Nil(t, err, "body read error")
	return body, rec.Code
}
