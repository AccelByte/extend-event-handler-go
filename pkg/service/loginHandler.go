// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	pb "extend-event-handler/pkg/pb/accelbyte-asyncapi/iam/account/v1"

	"extend-event-handler/pkg/common"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/entitlement"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
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
	entitlement platform.EntitlementService
}

func NewLoginHandler(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
) *LoginHandler {
	return &LoginHandler{
		entitlement: platform.EntitlementService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
	}
}

func (o *LoginHandler) grantEntitlement(userID string, itemID string, count int32) error {
	namespace := common.GetEnv("AB_NAMESPACE", "accelbyte")
	entitlementInfo, err := o.entitlement.GrantUserEntitlementShort(&entitlement.GrantUserEntitlementParams{
		Namespace: namespace,
		UserID:    userID,
		Body: []*platformclientmodels.EntitlementGrant{
			{
				ItemID:        &itemID,
				Quantity:      &count,
				Source:        platformclientmodels.EntitlementGrantSourceREWARD,
				ItemNamespace: &namespace,
			},
		},
	})
	if err != nil {
		return err
	}
	if len(entitlementInfo) <= 0 {
		return status.Errorf(codes.Internal, "could not grant item to user")
	}

	return nil
}

func (o *LoginHandler) OnMessage(ctx context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
	scope := common.GetScopeFromContext(ctx, "LoginHandler.OnMessage")
	defer scope.Finish()

	if itemIdToGrant == "" {
		return &emptypb.Empty{}, status.Errorf(
			codes.Internal, "Required envar ITEM_ID_TO_GRANT is not configured")
	}
	err := o.grantEntitlement(msg.UserId, itemIdToGrant, 1)

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "failed to grant entitlement: %v", err)
	}
	logrus.Infof("received a message: %v", msg)

	return &emptypb.Empty{}, nil
}
