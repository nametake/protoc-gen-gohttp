package main

import (
	"context"
)

var _ MessagingServer = (*Messaging)(nil)

type Messaging struct{}

func (m *Messaging) GetMessage(ctx context.Context, req *GetMessageRequest) (*GetMessageResponse, error) {
	return &GetMessageResponse{
		MessageId: req.MessageId,
		Message:   "Hello World!",
		Tags:      req.Tags,
	}, nil
}
