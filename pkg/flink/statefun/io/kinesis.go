package io

import (
	"github.com/sjwiesman/statefun-go/pkg/flink/statefun/internal"
	"github.com/sjwiesman/statefun-go/pkg/flink/statefun/internal/errors"
	"github.com/sjwiesman/statefun-go/pkg/flink/statefun/internal/messages"
	"google.golang.org/protobuf/proto"
)

// Egress message that will be
// written out to AWS Kinesis
type KinesisRecord struct {
	// Target AWS Kinesis stream to write to.
	Stream string

	// Partition key to use when writing
	// the record to AWS Kinesis.
	PartitionKey string

	// Optional explicit hash key to use
	// when writing the record to
	// the stream.
	ExplicitHashKey string

	// The message to write out to
	// the target stream.
	Value proto.Message
}

// Transforms a KinesisRecord into a Message that can
// be sent to an egress.
func (record *KinesisRecord) ToMessage() (proto.Message, error) {
	if record.Stream == "" {
		return nil, errors.New("cannot send a message to an empty stream")
	}

	marshalled, err := internal.Marshall(record.Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshall message meant for kinesis stream %s", record.Stream)
	}

	var bytes []byte

	if marshalled != nil {
		bytes, err = proto.Marshal(marshalled)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize message meant for kinesis stream %s", record.Stream)
		}
	}

	return &messages.KinesisEgressRecord{
		PartitionKey:    record.PartitionKey,
		ValueBytes:      bytes,
		Stream:          record.Stream,
		ExplicitHashKey: record.ExplicitHashKey,
	}, nil
}
