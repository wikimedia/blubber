package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func TestBlubberoidSpecification(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?spec", nil)

	blubberoid(rec, req)

	resp := rec.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	if assert.NotEmpty(t, body) {
		loader := openapi3.NewLoader()
		spec, err := loader.LoadFromData(body)

		if assert.NoError(t, err) {
			assert.NoError(t, spec.Validate(context.Background()))
		}
	}
}

func TestBlubberoidYAMLRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/test", strings.NewReader(`---
    version: v4
    base: foo
    variants:
      test: {}`))
	req.Header.Set("Content-Type", "application/yaml")

	blubberoid(rec, req)

	resp := rec.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "FROM foo")
	assert.Contains(t, string(body), `LABEL blubber.variant="test"`)
}

func TestBlubberoidJSONRequest(t *testing.T) {
	t.Run("valid JSON syntax", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/test", strings.NewReader(`{
			"version": "v4",
			"base": "foo",
			"variants": {
				"test": {}
			}
		}`))
		req.Header.Set("Content-Type", "application/json")

		blubberoid(rec, req)

		resp := rec.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
		assert.Contains(t, string(body), "FROM foo")
		assert.Contains(t, string(body), `LABEL blubber.variant="test"`)
	})

	t.Run("invalid JSON syntax", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/v1/test", strings.NewReader(`{
			version: "v4",
			base: "foo",
			variants: {
				test: {},
			},
		}`))
		req.Header.Set("Content-Type", "application/json")

		blubberoid(rec, req)

		resp := rec.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, string(body), "Failed to read 'application/json' config from request body. Error: invalid character 'v' looking for beginning of object key string\n")
	})
}

func TestBlubberoidUnsupportedMediaType(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/test", strings.NewReader(``))
	req.Header.Set("Content-Type", "application/foo")

	blubberoid(rec, req)

	resp := rec.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	assert.Equal(t, string(body), "'application/foo' media type is not supported\n")
}
