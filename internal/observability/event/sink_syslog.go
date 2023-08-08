// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package event

import (
	"context"
	"fmt"

	gsyslog "github.com/hashicorp/go-syslog"

	"github.com/hashicorp/eventlogger"
)

var (
	_ eventlogger.Closer = (*SyslogSink)(nil)
	_ eventlogger.Node   = (*SyslogSink)(nil)
)

// SyslogSink is a sink node which handles writing events to syslog.
type SyslogSink struct {
	requiredFormat string
	logger         gsyslog.Syslogger
	telemetryChan  chan<- map[string]any
}

// NewSyslogSink should be used to create a new SyslogSink.
// Accepted options: WithFacility, WithTag,  WithChannel.
func NewSyslogSink(format string, opt ...Option) (*SyslogSink, error) {
	const op = "event.NewSyslogSink"

	opts, err := getOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("%s: error applying options: %w", op, err)
	}

	logger, err := gsyslog.NewLogger(gsyslog.LOG_INFO, opts.withFacility, opts.withTag)
	if err != nil {
		return nil, fmt.Errorf("%s: error creating syslogger: %w", op, err)
	}

	return &SyslogSink{
		requiredFormat: format,
		logger:         logger,
		telemetryChan:  opts.withChannel,
	}, nil
}

// Process handles writing the event to the syslog.
func (s *SyslogSink) Process(ctx context.Context, e *eventlogger.Event) (*eventlogger.Event, error) {
	const op = "event.(SyslogSink).Process"

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if e == nil {
		return nil, fmt.Errorf("%s: event is nil: %w", op, ErrInvalidParameter)
	}

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

	formatted, found := e.Format(s.requiredFormat)
	if !found {
		return nil, fmt.Errorf("%s: unable to retrieve event formatted as %q", op, s.requiredFormat)
	}

	_, err := s.logger.Write(formatted)
	if err != nil {
		return nil, fmt.Errorf("%s: error writing to syslog: %w", op, err)
	}

	// update telemetry and return nil for the event to indicate the pipeline is complete.
	m["success"] = true
	return nil, nil
}

// Reopen is a no-op for a syslog sink.
func (_ *SyslogSink) Reopen() error {
	return nil
}

// Type describes the type of this node (sink).
func (_ *SyslogSink) Type() eventlogger.NodeType {
	return eventlogger.NodeTypeSink
}

// Close can be called by the eventlogger.Broker when nodes are removed to ensure
// they close any resources they are holding.
func (s *SyslogSink) Close(_ context.Context) error {
	if s.telemetryChan != nil {
		close(s.telemetryChan)
	}

	return nil
}
