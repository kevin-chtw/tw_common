package utils

import (
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TypeUrl(src proto.Message) string {
	any, err := anypb.New(src)
	if err != nil {
		logger.Log.Error(err)
		return ""
	}

	return any.GetTypeUrl()
}

func ToAny(ack proto.Message) *anypb.Any {
	data, err := anypb.New(ack)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}
	return data
}
