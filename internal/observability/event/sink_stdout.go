// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/eventlogger"
)

var _ eventlogger.Node = (*StdoutSink)(nil)

// StdoutSink is structure that implements the eventlogger.Node interface
// as a Sink node that writes the events to the standard output stream.
type StdoutSink struct {
	requiredFormat string
	telemetryChan  chan<- map[string]any // TODO: PW: strong type?
}

// NewStdoutSinkNode creates a new StdoutSink that will persist the events
// it processes using the specified expected format.
// Accepted options: WithChannel.
func NewStdoutSinkNode(format string, opt ...Option) (*StdoutSink, error) {
	const op = "event.NewStdoutSinkNode"

	opts, err := getOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("%s: error applying options: %w", op, err)
	}

	return &StdoutSink{
		requiredFormat: format,
		telemetryChan:  opts.withChannel,
	}, nil
}

// Process persists the provided eventlogger.Event to the standard output stream.
func (s *StdoutSink) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	const op = "event.(StdoutSink).Process"

	// Telemetry data
	m := map[string]any{
		"success": false,
		"created": e.CreatedAt,
	}

	// Ensure we emit telemetry before returning.
	defer func() {
		if s.telemetryChan != nil {
			s.telemetryChan <- m
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if e == nil {
		return nil, fmt.Errorf("%s: event is nil: %w", op, ErrInvalidParameter)
	}

	formattedBytes, found := e.Format(s.requiredFormat)
	if !found {
		return nil, fmt.Errorf("%s: unable to retrieve event formatted as %q", op, s.requiredFormat)
	}

	_, err := os.Stdout.Write(formattedBytes)
	if err != nil {
		return nil, fmt.Errorf("%s: error writing to stdout: %w", op, err)
	}

	// update telemetry and return nil for the event to indicate the pipeline is complete.
	m["success"] = true
	return nil, nil
}

// Reopen is a no-op for the StdoutSink type.
func (s *StdoutSink) Reopen() error {
	return nil
}

// Type returns the eventlogger.NodeTypeSink constant.
func (s *StdoutSink) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeSink
}
