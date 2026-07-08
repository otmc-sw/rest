/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
**/

package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	restcontext "github.com/otmc-sw/rest/context"
	"github.com/otmc-sw/rest/errors"
)

type nextFunc func(ctx restcontext.Context) error

func RequestID(headerKey string) func(restcontext.Context, nextFunc) error {
	if headerKey == "" {
		headerKey = "X-Request-ID"
	}
	return func(ctx restcontext.Context, next nextFunc) error {
		id := generateID()
		ctx.SetHeader(headerKey, id)
		return next(ctx)
	}
}

func Logger() func(restcontext.Context, nextFunc) error {
	return func(ctx restcontext.Context, next nextFunc) error {
		start := time.Now()
		err := next(ctx)
		log.Printf("[rest] %s %s %s", ctx.Method(), ctx.Path(), time.Since(start))
		return err
	}
}

func Recover() func(restcontext.Context, nextFunc) error {
	return func(ctx restcontext.Context, next nextFunc) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[rest] panic recovered: %v", r)
				sendErr := errors.New().
					InternalError().
					Summary("internal server error").
					Detail(r).
					Send(ctx)
				if sendErr != nil {
					err = sendErr
				}
			}
		}()
		return next(ctx)
	}
}

func Timeout(d time.Duration) func(restcontext.Context, nextFunc) error {
	return func(ctx restcontext.Context, next nextFunc) error {
		derived, cancel := context.WithTimeout(ctx.Context(), d)
		defer cancel()
		if setter, ok := ctx.(interface {
			SetContext(context.Context)
		}); ok {
			setter.SetContext(derived)
		}
		return next(ctx)
	}
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}