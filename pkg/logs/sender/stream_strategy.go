// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package sender

import (
	"github.com/DataDog/datadog-agent/pkg/logs/message"
	"github.com/DataDog/datadog-agent/pkg/util/compression"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// streamStrategy is a Strategy that creates one Payload for each Message, containing
// that Message's Content. This is used for TCP destinations, which stream the output
// without batching multiple messages together.
type streamStrategy struct {
	inputChan   chan *message.Message
	outputChan  chan *message.Payload
	compression compression.Compressor
	done        chan struct{}
}

// NewStreamStrategy creates a new stream strategy
func NewStreamStrategy(inputChan chan *message.Message, outputChan chan *message.Payload, compression compression.Compressor) Strategy {
	return &streamStrategy{
		inputChan:   inputChan,
		outputChan:  outputChan,
		compression: compression,
		done:        make(chan struct{}),
	}
}

// Send sends one message at a time and forwards them to the next stage of the pipeline.
func (s *streamStrategy) Start() {
	go func() {
		for msg := range s.inputChan {
			if msg.Origin != nil {
				msg.Origin.LogSource.LatencyStats.Add(msg.GetLatency())
			}

			encodedPayload, err := s.compression.Compress(msg.GetContent())
			if err != nil {
				log.Warn("Encoding failed - dropping payload", err)
				return
			}

			unencodedSize := len(msg.GetContent())

			// Split the metadata from the message content to avoid holding the entire message in memory
			meta := msg.MessageMetadata
			s.outputChan <- message.NewPayload([]*message.MessageMetadata{&meta}, encodedPayload, s.compression.ContentEncoding(), unencodedSize)
		}
		s.done <- struct{}{}
	}()
}

// Stop stops the strategy
func (s *streamStrategy) Stop() {
	close(s.inputChan)
	<-s.done
}
