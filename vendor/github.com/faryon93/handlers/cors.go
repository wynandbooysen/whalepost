package handlers

// handlers
// Copyright (C) 2018 Maximilian Pachl

// MIT License
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"net/url"
	"strings"
	"net/http"

	"github.com/gorilla/handlers"

	"github.com/faryon93/handlers/opt"
)

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// CORS creates a gorilla CORS adapter.
// The age parameter is in seconds.
func CORS(age int, origins ...string) Adapter {
	methods := handlers.AllowedMethods([]string{
		"GET", "HEAD", "POST", "PATCH", "DELETE",
	})
	headers := handlers.AllowedHeaders([]string{
		"Content-Type", "Authorization", "X-Body-Signature",
	})
    exposed := handlers.ExposedHeaders([]string{
        "X-Body-Signature",
    })
	validator := handlers.AllowedOriginValidator(func(requestOrigin string) bool {
		return isOriginValid(origins, requestOrigin)
	})

	return handlers.CORS(methods, headers, validator, exposed,
		handlers.AllowCredentials(), handlers.MaxAge(age))
}

// RestrictOrigin prevents further processing if the request origin
// is not on the origins list.
func RestrictOrigin(origins []string, opts ...interface{}) Adapter {
	httpError := opt.GetErrorHandler(opts)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && !isOriginValid(origins, origin) {
				httpError(w, "origin restriction", http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

}

// ---------------------------------------------------------------------------------------
//  private functions
// ---------------------------------------------------------------------------------------

// Returns true if the given requestOrigin is on the origins list.
func isOriginValid(origins []string, requestOrigin string) bool {
	originUrl, err := url.Parse(requestOrigin)
	if err != nil {
		return false
	}

	for _, origin := range origins {
		if strings.HasSuffix(originUrl.Hostname(), origin) {
			return true
		}
	}

	return false
}
