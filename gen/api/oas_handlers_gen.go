// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-faster/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/otelogen"
)

// handleAuthorizeGetRequest handles GET /authorize operation.
//
// GET /authorize
func (s *Server) handleAuthorizeGetRequest(args [0]string, argsEscaped bool, w http.ResponseWriter, r *http.Request) {
	otelAttrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String("GET"),
		semconv.HTTPRouteKey.String("/authorize"),
	}

	// Start a span for this request.
	ctx, span := s.cfg.Tracer.Start(r.Context(), "AuthorizeGet",
		trace.WithAttributes(otelAttrs...),
		serverSpanKind,
	)
	defer span.End()

	// Add Labeler to context.
	labeler := &Labeler{attrs: otelAttrs}
	ctx = contextWithLabeler(ctx, labeler)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		elapsedDuration := time.Since(startTime)
		attrOpt := metric.WithAttributeSet(labeler.AttributeSet())

		// Increment request counter.
		s.requests.Add(ctx, 1, attrOpt)

		// Use floating point division here for higher precision (instead of Millisecond method).
		s.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), attrOpt)
	}()

	var (
		recordError = func(stage string, err error) {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			s.errors.Add(ctx, 1, metric.WithAttributeSet(labeler.AttributeSet()))
		}
		err          error
		opErrContext = ogenerrors.OperationContext{
			Name: "AuthorizeGet",
			ID:   "",
		}
	)
	params, err := decodeAuthorizeGetParams(args, argsEscaped, r)
	if err != nil {
		err = &ogenerrors.DecodeParamsError{
			OperationContext: opErrContext,
			Err:              err,
		}
		defer recordError("DecodeParams", err)
		s.cfg.ErrorHandler(ctx, w, r, err)
		return
	}

	var response AuthorizeGetRes
	if m := s.cfg.Middleware; m != nil {
		mreq := middleware.Request{
			Context:          ctx,
			OperationName:    "AuthorizeGet",
			OperationSummary: "",
			OperationID:      "",
			Body:             nil,
			Params: middleware.Parameters{
				{
					Name: "response_type",
					In:   "query",
				}: params.ResponseType,
				{
					Name: "client_id",
					In:   "query",
				}: params.ClientID,
				{
					Name: "scope",
					In:   "query",
				}: params.Scope,
				{
					Name: "redirect_uri",
					In:   "query",
				}: params.RedirectURI,
				{
					Name: "state",
					In:   "query",
				}: params.State,
				{
					Name: "nonce",
					In:   "query",
				}: params.Nonce,
			},
			Raw: r,
		}

		type (
			Request  = struct{}
			Params   = AuthorizeGetParams
			Response = AuthorizeGetRes
		)
		response, err = middleware.HookMiddleware[
			Request,
			Params,
			Response,
		](
			m,
			mreq,
			unpackAuthorizeGetParams,
			func(ctx context.Context, request Request, params Params) (response Response, err error) {
				response, err = s.h.AuthorizeGet(ctx, params)
				return response, err
			},
		)
	} else {
		response, err = s.h.AuthorizeGet(ctx, params)
	}
	if err != nil {
		if errRes, ok := errors.Into[*ErrRespStatusCode](err); ok {
			if err := encodeErrorResponse(errRes, w, span); err != nil {
				defer recordError("Internal", err)
			}
			return
		}
		if errors.Is(err, ht.ErrNotImplemented) {
			s.cfg.ErrorHandler(ctx, w, r, err)
			return
		}
		if err := encodeErrorResponse(s.h.NewError(ctx, err), w, span); err != nil {
			defer recordError("Internal", err)
		}
		return
	}

	if err := encodeAuthorizeGetResponse(response, w, span); err != nil {
		defer recordError("EncodeResponse", err)
		if !errors.Is(err, ht.ErrInternalServerErrorResponse) {
			s.cfg.ErrorHandler(ctx, w, r, err)
		}
		return
	}
}

// handleIndexRequest handles Index operation.
//
// Root URL.
//
// GET /
func (s *Server) handleIndexRequest(args [0]string, argsEscaped bool, w http.ResponseWriter, r *http.Request) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("Index"),
		semconv.HTTPRequestMethodKey.String("GET"),
		semconv.HTTPRouteKey.String("/"),
	}

	// Start a span for this request.
	ctx, span := s.cfg.Tracer.Start(r.Context(), "Index",
		trace.WithAttributes(otelAttrs...),
		serverSpanKind,
	)
	defer span.End()

	// Add Labeler to context.
	labeler := &Labeler{attrs: otelAttrs}
	ctx = contextWithLabeler(ctx, labeler)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		elapsedDuration := time.Since(startTime)
		attrOpt := metric.WithAttributeSet(labeler.AttributeSet())

		// Increment request counter.
		s.requests.Add(ctx, 1, attrOpt)

		// Use floating point division here for higher precision (instead of Millisecond method).
		s.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), attrOpt)
	}()

	var (
		recordError = func(stage string, err error) {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			s.errors.Add(ctx, 1, metric.WithAttributeSet(labeler.AttributeSet()))
		}
		err error
	)

	var response IndexOK
	if m := s.cfg.Middleware; m != nil {
		mreq := middleware.Request{
			Context:          ctx,
			OperationName:    "Index",
			OperationSummary: "",
			OperationID:      "Index",
			Body:             nil,
			Params:           middleware.Parameters{},
			Raw:              r,
		}

		type (
			Request  = struct{}
			Params   = struct{}
			Response = IndexOK
		)
		response, err = middleware.HookMiddleware[
			Request,
			Params,
			Response,
		](
			m,
			mreq,
			nil,
			func(ctx context.Context, request Request, params Params) (response Response, err error) {
				response, err = s.h.Index(ctx)
				return response, err
			},
		)
	} else {
		response, err = s.h.Index(ctx)
	}
	if err != nil {
		if errRes, ok := errors.Into[*ErrRespStatusCode](err); ok {
			if err := encodeErrorResponse(errRes, w, span); err != nil {
				defer recordError("Internal", err)
			}
			return
		}
		if errors.Is(err, ht.ErrNotImplemented) {
			s.cfg.ErrorHandler(ctx, w, r, err)
			return
		}
		if err := encodeErrorResponse(s.h.NewError(ctx, err), w, span); err != nil {
			defer recordError("Internal", err)
		}
		return
	}

	if err := encodeIndexResponse(response, w, span); err != nil {
		defer recordError("EncodeResponse", err)
		if !errors.Is(err, ht.ErrInternalServerErrorResponse) {
			s.cfg.ErrorHandler(ctx, w, r, err)
		}
		return
	}
}

// handleJwksRequest handles jwks operation.
//
// Json Web Keyset.
//
// GET /.well-known/jwks.json
func (s *Server) handleJwksRequest(args [0]string, argsEscaped bool, w http.ResponseWriter, r *http.Request) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("jwks"),
		semconv.HTTPRequestMethodKey.String("GET"),
		semconv.HTTPRouteKey.String("/.well-known/jwks.json"),
	}

	// Start a span for this request.
	ctx, span := s.cfg.Tracer.Start(r.Context(), "Jwks",
		trace.WithAttributes(otelAttrs...),
		serverSpanKind,
	)
	defer span.End()

	// Add Labeler to context.
	labeler := &Labeler{attrs: otelAttrs}
	ctx = contextWithLabeler(ctx, labeler)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		elapsedDuration := time.Since(startTime)
		attrOpt := metric.WithAttributeSet(labeler.AttributeSet())

		// Increment request counter.
		s.requests.Add(ctx, 1, attrOpt)

		// Use floating point division here for higher precision (instead of Millisecond method).
		s.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), attrOpt)
	}()

	var (
		recordError = func(stage string, err error) {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			s.errors.Add(ctx, 1, metric.WithAttributeSet(labeler.AttributeSet()))
		}
		err error
	)

	var response *JWKSetResponse
	if m := s.cfg.Middleware; m != nil {
		mreq := middleware.Request{
			Context:          ctx,
			OperationName:    "Jwks",
			OperationSummary: "",
			OperationID:      "jwks",
			Body:             nil,
			Params:           middleware.Parameters{},
			Raw:              r,
		}

		type (
			Request  = struct{}
			Params   = struct{}
			Response = *JWKSetResponse
		)
		response, err = middleware.HookMiddleware[
			Request,
			Params,
			Response,
		](
			m,
			mreq,
			nil,
			func(ctx context.Context, request Request, params Params) (response Response, err error) {
				response, err = s.h.Jwks(ctx)
				return response, err
			},
		)
	} else {
		response, err = s.h.Jwks(ctx)
	}
	if err != nil {
		if errRes, ok := errors.Into[*ErrRespStatusCode](err); ok {
			if err := encodeErrorResponse(errRes, w, span); err != nil {
				defer recordError("Internal", err)
			}
			return
		}
		if errors.Is(err, ht.ErrNotImplemented) {
			s.cfg.ErrorHandler(ctx, w, r, err)
			return
		}
		if err := encodeErrorResponse(s.h.NewError(ctx, err), w, span); err != nil {
			defer recordError("Internal", err)
		}
		return
	}

	if err := encodeJwksResponse(response, w, span); err != nil {
		defer recordError("EncodeResponse", err)
		if !errors.Is(err, ht.ErrInternalServerErrorResponse) {
			s.cfg.ErrorHandler(ctx, w, r, err)
		}
		return
	}
}

// handleOpenIdConfigurationRequest handles openIdConfiguration operation.
//
// (Partial) OIDC config.
//
// GET /.well-known/openid-configuration
func (s *Server) handleOpenIdConfigurationRequest(args [0]string, argsEscaped bool, w http.ResponseWriter, r *http.Request) {
	otelAttrs := []attribute.KeyValue{
		otelogen.OperationID("openIdConfiguration"),
		semconv.HTTPRequestMethodKey.String("GET"),
		semconv.HTTPRouteKey.String("/.well-known/openid-configuration"),
	}

	// Start a span for this request.
	ctx, span := s.cfg.Tracer.Start(r.Context(), "OpenIdConfiguration",
		trace.WithAttributes(otelAttrs...),
		serverSpanKind,
	)
	defer span.End()

	// Add Labeler to context.
	labeler := &Labeler{attrs: otelAttrs}
	ctx = contextWithLabeler(ctx, labeler)

	// Run stopwatch.
	startTime := time.Now()
	defer func() {
		elapsedDuration := time.Since(startTime)
		attrOpt := metric.WithAttributeSet(labeler.AttributeSet())

		// Increment request counter.
		s.requests.Add(ctx, 1, attrOpt)

		// Use floating point division here for higher precision (instead of Millisecond method).
		s.duration.Record(ctx, float64(float64(elapsedDuration)/float64(time.Millisecond)), attrOpt)
	}()

	var (
		recordError = func(stage string, err error) {
			span.RecordError(err)
			span.SetStatus(codes.Error, stage)
			s.errors.Add(ctx, 1, metric.WithAttributeSet(labeler.AttributeSet()))
		}
		err error
	)

	var response *OpenIDProviderMetadataResponse
	if m := s.cfg.Middleware; m != nil {
		mreq := middleware.Request{
			Context:          ctx,
			OperationName:    "OpenIdConfiguration",
			OperationSummary: "",
			OperationID:      "openIdConfiguration",
			Body:             nil,
			Params:           middleware.Parameters{},
			Raw:              r,
		}

		type (
			Request  = struct{}
			Params   = struct{}
			Response = *OpenIDProviderMetadataResponse
		)
		response, err = middleware.HookMiddleware[
			Request,
			Params,
			Response,
		](
			m,
			mreq,
			nil,
			func(ctx context.Context, request Request, params Params) (response Response, err error) {
				response, err = s.h.OpenIdConfiguration(ctx)
				return response, err
			},
		)
	} else {
		response, err = s.h.OpenIdConfiguration(ctx)
	}
	if err != nil {
		if errRes, ok := errors.Into[*ErrRespStatusCode](err); ok {
			if err := encodeErrorResponse(errRes, w, span); err != nil {
				defer recordError("Internal", err)
			}
			return
		}
		if errors.Is(err, ht.ErrNotImplemented) {
			s.cfg.ErrorHandler(ctx, w, r, err)
			return
		}
		if err := encodeErrorResponse(s.h.NewError(ctx, err), w, span); err != nil {
			defer recordError("Internal", err)
		}
		return
	}

	if err := encodeOpenIdConfigurationResponse(response, w, span); err != nil {
		defer recordError("EncodeResponse", err)
		if !errors.Is(err, ht.ErrInternalServerErrorResponse) {
			s.cfg.ErrorHandler(ctx, w, r, err)
		}
		return
	}
}