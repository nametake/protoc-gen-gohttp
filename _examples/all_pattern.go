package main

import (
	"context"
)

var _ AllPatternServer = (*AllPattern)(nil)

type AllPattern struct{}

func (a *AllPattern) AllPattern(ctx context.Context, msg *AllPatternMessage) (*AllPatternMessage, error) {
	return msg, nil
}
