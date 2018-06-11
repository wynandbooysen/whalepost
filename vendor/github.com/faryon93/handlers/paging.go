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
    "strconv"
    "context"

    "github.com/faryon93/handlers/opt"
)

// ---------------------------------------------------------------------------------------
//  constants
// ---------------------------------------------------------------------------------------

const (
    ctxPageSkip = "page_skip"
    ctxPageLimit = "page_limit"
)


// ---------------------------------------------------------------------------------------
//  public functions
// ---------------------------------------------------------------------------------------

// Paged parses the limit and skip parameter and stores it to the requests context.
func Paged(defaultLimit string, opts ...interface{}) Adapter {
    httpError := opt.GetErrorHandler(opts)

    return func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            skipStr := r.URL.Query().Get("skip")
            limitStr := r.URL.Query().Get("limit")

            if skipStr == "" {
                skipStr = "0"
            }

            if limitStr == "" {
                limitStr = defaultLimit
            }

            skip, err := strconv.Atoi(skipStr)
            if err != nil {
                httpError(w, "paging error: " + err.Error(), http.StatusBadRequest)
                return
            }

            limit, err := strconv.Atoi(limitStr)
            if err != nil {
                httpError(w, "paging error: " + err.Error(), http.StatusBadRequest)
                return
            }

            if skip < 0 || limit < 1 {
                httpError(w,"invalid paging value", http.StatusBadRequest)
                return
            }

            ctx := context.WithValue(r.Context(), ctxPageSkip, skip)
            ctx = context.WithValue(ctx, ctxPageLimit, limit)
            h.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GetPageSkip returns the element offset aka skip of the Paged() Adapater.
// To use this funktion the Paged() adapter need to be chained in.
func GetPageSkip(r *http.Request) int {
    return r.Context().Value(ctxPageSkip).(int)
}

// GetPageLimit returns the number of elements per page aka limit of the Paged() Adapater.
// To use this funktion the Paged() adapter need to be chained in.
func GetPageLimit(r *http.Request) int {
    return r.Context().Value(ctxPageLimit).(int)
}