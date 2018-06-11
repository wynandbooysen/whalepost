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
	"net/http"

	"github.com/faryon93/util"
	"github.com/haisum/recaptcha"

	"github.com/faryon93/handlers/opt"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

type recaptchaInput struct {
	Response string `json:"g-recaptcha-response" schema:"g-recaptcha-response"`
}

// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// Recaptcha verifies the given recaptcha response.
// If the verification failes the processing of the request is canceled.
// If the key parameter is empty, the captcha check is bypassed.
func Recaptcha(key string, opts ...interface{}) Adapter {
	captcha := recaptcha.R{Secret: key}
	httpError := opt.GetErrorHandler(opts)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				// save the body to preserve it for future use
				body := util.SaveRequestBody(r)

				// we want the recaptcha response field
				var form recaptchaInput
				err := util.ParseBody(r, &form)
				if err != nil {
					httpError(w, "mailformed request: "+err.Error(), http.StatusBadRequest)
					return
				}

				// verfiy the captcha with the captcha server
				ok := captcha.VerifyResponse(form.Response)
				if !ok {
					httpError(w, "invalid recaptcha response", http.StatusBadRequest)
					return
				}

				body.Restore(r)
			}

			h.ServeHTTP(w, r)
		})
	}
}
