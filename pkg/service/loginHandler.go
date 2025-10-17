// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"context"
	pb "extend-event-handler/pkg/pb/accelbyte-asyncapi/iam/account/v1"

	"extend-event-handler/pkg/common"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
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
	fulfillment platform.FulfillmentService
}

func NewLoginHandler(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
) *LoginHandler {
	return &LoginHandler{
		fulfillment: platform.FulfillmentService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
	}
}

func (o *LoginHandler) grantEntitlement(userID string, itemID string, count int32) error {
	namespace := common.GetEnv("AB_NAMESPACE", "accelbyte")
	fulfillmentResponse, err := o.fulfillment.FulfillItemShort(&fulfillment.FulfillItemParams{
		Namespace: namespace,
		UserID:    userID,
		Body: &platformclientmodels.FulfillmentRequest{
			ItemID:   itemID,
			Quantity: &count,
			Source:   platformclientmodels.EntitlementGrantSourceREWARD,
		},
	})
	if err != nil {
		return err
	}
	if fulfillmentResponse == nil || fulfillmentResponse.EntitlementSummaries == nil || len(fulfillmentResponse.EntitlementSummaries) <= 0 {
		return status.Errorf(codes.Internal, "could not grant item to user")
	}

	return nil
}

func (o *LoginHandler) OnMessage(ctx context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
	scope := common.GetScopeFromContext(ctx, "LoginHandler.OnMessage")
	defer scope.Finish()

	logrus.Infof("received an event: %v", msg)

	if itemIdToGrant == "" {
		return &emptypb.Empty{}, status.Errorf(
			codes.Internal, "Required envar ITEM_ID_TO_GRANT is not configured")
	}
	err := o.grantEntitlement(msg.UserId, itemIdToGrant, 1)

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "failed to grant entitlement: %v", err)
	}

	return &emptypb.Empty{}, nil
}
