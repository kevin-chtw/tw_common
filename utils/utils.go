package utils

import (
	"context"

	"github.com/topfreegames/pitaya/v3/pkg/constants"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
	"github.com/topfreegames/pitaya/v3/pkg/session"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	NetType     = "net_type"
	JsonMarshal = protojson.MarshalOptions{UseProtoNames: true}
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

func IsWebsocket(ctx context.Context) bool {
	sessionVal := ctx.Value(constants.SessionCtxKey)
	if sessionVal == nil {
		return false
	}
	s := sessionVal.(session.Session)
	netType := s.String(NetType)
	return netType == "ws"
}

func Unmarshal(ctx context.Context, payLoads []byte, msg proto.Message) error {
	if IsWebsocket(ctx) {
		return protojson.Unmarshal(payLoads, msg)
	} else {
		return proto.Unmarshal(payLoads, msg)
	}
}

func Marshal(ctx context.Context, out proto.Message) ([]byte, error) {
	if IsWebsocket(ctx) {
		return JsonMarshal.Marshal(out)
	} else {
		return proto.Marshal(out)
	}
}
