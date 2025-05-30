// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022-present Datadog, Inc.

package daemon

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/DataDog/datadog-agent/pkg/serverless/invocationlifecycle"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Hello is a route called by the Datadog Lambda Library when it starts.
// It is used to detect the Datadog Lambda Library in the environment.
type Hello struct {
	daemon *Daemon
}

//nolint:revive // TODO(SERV) Fix revive linter
func (h *Hello) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	log.Debug("Hit on the serverless.Hello route.")
	h.daemon.LambdaLibraryStateLock.Lock()
	defer h.daemon.LambdaLibraryStateLock.Unlock()
	h.daemon.LambdaLibraryDetected = true
}

// Flush is a route called by the Datadog Lambda Library when the runtime is done handling an invocation.
// It is no longer used, but the route is maintained for backwards compatibility.
type Flush struct {
	daemon *Daemon
}

//nolint:revive // TODO(SERV) Fix revive linter
func (f *Flush) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	log.Debug("Hit on the serverless.Flush route.")
	if os.Getenv(LocalTestEnvVar) == "true" || os.Getenv(LocalTestEnvVar) == "1" {
		// used only for testing purpose as the Logs API is not supported by the Lambda Emulator
		// thus we canot get the REPORT log line telling that the invocation is finished
		f.daemon.HandleRuntimeDone()
	}
}

// StartInvocation is a route that can be called at the beginning of an invocation to enable
// the invocation lifecyle feature without the use of the proxy.
type StartInvocation struct {
	daemon *Daemon
}

func (s *StartInvocation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("Hit on the serverless.StartInvocation route.")
	s.daemon.SetExecutionSpanIncomplete(true)
	startTime := time.Now()
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("Could not read StartInvocation request body")
		http.Error(w, "Could not read StartInvocation request body", 400)
		return
	}
	startDetails := &invocationlifecycle.InvocationStartDetails{
		StartTime:             startTime,
		InvokeEventRawPayload: reqBody,
		InvokeEventHeaders:    r.Header,
		InvokedFunctionARN:    s.daemon.ExecutionContext.GetCurrentState().ARN,
	}

	s.daemon.InvocationProcessor.OnInvokeStart(startDetails)

	if s.daemon.InvocationProcessor.GetExecutionInfo().TraceID == 0 {
		log.Debug("no context has been found, the tracer will be responsible for initializing the context")
	} else {
		log.Debug("a context has been found, sending the context to the tracer")
		w.Header().Set(invocationlifecycle.TraceIDHeader, fmt.Sprintf("%v", s.daemon.InvocationProcessor.GetExecutionInfo().TraceID))
		w.Header().Set(invocationlifecycle.SamplingPriorityHeader, fmt.Sprintf("%v", s.daemon.InvocationProcessor.GetExecutionInfo().SamplingPriority))
		if s.daemon.InvocationProcessor.GetExecutionInfo().TraceIDUpper64Hex != "" {
			w.Header().Set(invocationlifecycle.TraceTagsHeader, fmt.Sprintf("%s=%s", invocationlifecycle.Upper64BitsTag, s.daemon.InvocationProcessor.GetExecutionInfo().TraceIDUpper64Hex))
		}
	}
}

// EndInvocation is a route that can be called at the end of an invocation to enable
// the invocation lifecycle feature without the use of the proxy.
type EndInvocation struct {
	daemon *Daemon
}

func (e *EndInvocation) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("Hit on the serverless.EndInvocation route.")
	e.daemon.SetExecutionSpanIncomplete(false)
	endTime := time.Now()
	ecs := e.daemon.ExecutionContext.GetCurrentState()
	coldStartTags := e.daemon.ExecutionContext.GetColdStartTagsForRequestID(ecs.LastRequestID)
	responseBody, err := io.ReadAll(r.Body)
	if err != nil {
		err := log.Error("Could not read EndInvocation request body")
		http.Error(w, err.Error(), 400)
		return
	}

	errorMsg := r.Header.Get(invocationlifecycle.InvocationErrorMsgHeader)
	if decodedMsg, err := base64.StdEncoding.DecodeString(errorMsg); err != nil {
		log.Debug("Error message header may not be encoded, setting as is")
	} else {
		errorMsg = string(decodedMsg)
	}
	errorType := r.Header.Get(invocationlifecycle.InvocationErrorTypeHeader)
	if decodedType, err := base64.StdEncoding.DecodeString(errorType); err != nil {
		log.Debug("Error type header may not be encoded, setting as is")
	} else {
		errorType = string(decodedType)
	}
	errorStack := r.Header.Get(invocationlifecycle.InvocationErrorStackHeader)
	if decodedStack, err := base64.StdEncoding.DecodeString(errorStack); err != nil {
		log.Debug("Could not decode error stack header")
	} else {
		errorStack = string(decodedStack)
	}
	// If any error metadata is received, always mark the span as an error
	isError := r.Header.Get(invocationlifecycle.InvocationErrorHeader) == "true" || len(errorMsg) > 0 || len(errorType) > 0 || len(errorStack) > 0

	var endDetails = invocationlifecycle.InvocationEndDetails{
		EndTime:            endTime,
		IsError:            isError,
		RequestID:          ecs.LastRequestID,
		ResponseRawPayload: responseBody,
		ColdStart:          coldStartTags.IsColdStart,
		ProactiveInit:      coldStartTags.IsProactiveInit,
		Runtime:            ecs.Runtime,
		ErrorMsg:           errorMsg,
		ErrorType:          errorType,
		ErrorStack:         errorStack,
	}
	executionContext := e.daemon.InvocationProcessor.GetExecutionInfo()
	if executionContext.TraceID == 0 {
		log.Debug("no context has been found yet, injecting it now via headers from the tracer")
		invocationlifecycle.InjectContext(executionContext, r.Header)
	}
	invocationlifecycle.InjectSpanID(executionContext, r.Header)
	e.daemon.InvocationProcessor.OnInvokeEnd(&endDetails)
}

// TraceContext is a route called by tracer so it can retrieve the tracing context
type TraceContext struct {
	daemon *Daemon
}

//nolint:revive // TODO(SERV) Fix revive linter
func (tc *TraceContext) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	executionInfo := tc.daemon.InvocationProcessor.GetExecutionInfo()
	log.Debug("Hit on the serverless.TraceContext route.")
	w.Header().Set(invocationlifecycle.TraceIDHeader, fmt.Sprintf("%v", executionInfo.TraceID))
	w.Header().Set(invocationlifecycle.SpanIDHeader, fmt.Sprintf("%v", executionInfo.SpanID))
	w.Header().Set(invocationlifecycle.SamplingPriorityHeader, fmt.Sprintf("%v", executionInfo.SamplingPriority))
}
