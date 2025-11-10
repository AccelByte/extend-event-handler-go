// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	pb "extend-event-handler/pkg/pb/accelbyte-asyncapi/iam/account/v1"

	"extend-event-handler/pkg/common"

	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	itemIdToGrant = common.GetEnv("ITEM_ID_TO_GRANT", "")
)

type LoginHandler struct {
	pb.UnimplementedUserAuthenticationUserLoggedInServiceServer
	fulfillment platform.FulfillmentService
	namespace   string
}

func NewLoginHandler(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
	namespace string,
) *LoginHandler {
	return &LoginHandler{
		fulfillment: platform.FulfillmentService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
		namespace: namespace,
	}
}

func (o *LoginHandler) OnMessage(ctx context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
	scope := common.GetScopeFromContext(ctx, "LoginHandler.OnMessage")
	defer scope.Finish()

	logrus.Infof("received an event: %v", msg)

	err := grantEntitlement(o.fulfillment, o.namespace, msg.UserId, itemIdToGrant)

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "failed to grant entitlement: %v", err)
	}

	return &emptypb.Empty{}, nil
}
