package main

import (
	"context"
)

var _ MessagingHTTPService = (*Messaging)(nil)

type Messaging struct{}

func (m *Messaging) GetMessage(ctx context.Context, req *GetMessageRequest) (*GetMessageResponse, error) {
	return &GetMessageResponse{
		MessageId: req.MessageId,
		Message:   req.Message,
		Tags:      req.Tags,
	}, nil
}

func (m *Messaging) UpdateMessage(ctx context.Context, req *UpdateMessageRequest) (*UpdateMessageResponse, error) {
	return &UpdateMessageResponse{
		MessageId: req.MessageId,
		Sub: &SubMessage{
			Subfield: req.Sub.Subfield,
		},
		Message: req.Message,
	}, nil
}

func (m *Messaging) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*CreateMessageResponse, error) {
	return &CreateMessageResponse{
		MessageId: req.MessageId,
		Sub: &SubMessage{
			Subfield: req.Sub.Subfield,
		},
		Msg: &CreateMessageResponse_Message{
			Sub: &SubMessage{
				Subfield: req.Msg.Sub.Subfield,
			},
		},
		Opt: req.Opt,
	}, nil
}
