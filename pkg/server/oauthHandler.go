package server

import (
	"context"
	pb "github.com/001extend/extend-event-listener/pkg/pb/accelbyte-async-api/iam/oauth/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OauthHandler struct {
	pb.UnimplementedIAMServiceOAuthEventsServer
}

func NewOauthHandler() *OauthHandler {
	return &OauthHandler{}
}
func (o *OauthHandler) PublishToOauthRequestChannel(ctx context.Context, msg *pb.OauthRequestPublishMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishOauthRequestAuthorizedMessageToOauthRequestChannel(xy context.Context, msg *pb.OauthRequestAuthorizedMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishToOauthTokenChannel(ctx context.Context, msg *pb.OauthTokenPublishMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishOauthTokenGeneratedMessageToOauthTokenChannel(ctx context.Context, msg *pb.OauthTokenGeneratedMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishOauthTokenRevokedMessageToOauthTokenChannel(ctx context.Context, msg *pb.OauthTokenRevokedMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishToOauthThirdPartyRequestChannel(ctx context.Context, msg *pb.OauthThirdPartyRequestPublishMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishOauthThirdPartyRequestAuthorizedMessageToOauthThirdPartyRequestChannel(ctx context.Context, msg *pb.OauthThirdPartyRequestAuthorizedMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishToOauthThirdPartyTokenChannel(ctx context.Context, msg *pb.OauthThirdPartyTokenPublishMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
func (o *OauthHandler) PublishOauthThirdPartyTokenGeneratedMessageToOauthThirdPartyTokenChannel(ctx context.Context, msg *pb.OauthThirdPartyTokenGeneratedMessage) (*emptypb.Empty, error) {
	logrus.Infof("received a message: %v", msg)
	return &emptypb.Empty{}, nil
}
