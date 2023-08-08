// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"fmt"

	"github.com/hashicorp/eventlogger"
)

var (
	_ eventlogger.Closer = (*NoopSink)(nil)
	_ eventlogger.Node   = (*NoopSink)(nil)
)

// NoopSink is a sink node which ignores/discards everything.
type NoopSink struct {
	telemetryChan chan<- map[string]any // TODO: PW: strong type?
}

// NewNoopSink should be used to create a new NoopSink.
func NewNoopSink(opt ...Option) (*NoopSink, error) {
	const op = "event.NewNoopSink"

	opts, err := getOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("%s: error applying options: %w", op, err)
	}

	return &NoopSink{
		telemetryChan: opts.withChannel,
	}, nil
}

// Process is a no-op and always returns nil event and nil error.
func (s *NoopSink) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	// Ensure we emit telemetry before returning.
	defer func() {
		if s.telemetryChan != nil {
			s.telemetryChan <- map[string]any{
				"success": true,
				"created": e.CreatedAt,
			}
		}
	}()

	// return nil for the event to indicate the pipeline is complete.
	return nil, nil
}

// Reopen is a no-op and always returns nil.
func (s *NoopSink) Reopen() error {
	return nil
}

// Type describes the type of this node (sink).
func (s *NoopSink) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeSink
}

// Close can be called by the eventlogger.Broker when nodes are removed to ensure
// they close any resources they are holding.
func (s *NoopSink) Close(_ context.Context) error {
	if s.telemetryChan != nil {
		close(s.telemetryChan)
	}

	return nil
}
