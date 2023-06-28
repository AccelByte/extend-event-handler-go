// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package server

import (
	"context"
	pb "extend-event-listener/pkg/pb/accelbyte-asyncapi/iam/account/v1"

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
	itemIdToGrant = GetEnv("ITEM_ID_TO_GRANT", "")
)

type LoginListener struct {
	pb.UnimplementedUserAuthenticationUserLoggedInServiceServer
	entitlement platform.EntitlementService
}

func NewLoginListener(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
) *LoginListener {
	return &LoginListener{
		entitlement: platform.EntitlementService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
	}
}

func (o *LoginListener) grantEntitlement(userID string, itemID string, count int32) error {
	namespace := getNamespace()
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

func (o *LoginListener) OnMessage(_ context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
	if msg.Namespace != getNamespace() {
		return &emptypb.Empty{}, status.Errorf(
			codes.InvalidArgument,
			"user with namespace %s not belong to the configured namespace %s. Full message: %v",
			msg.Namespace, getNamespace(), msg)
	}
	err := o.grantEntitlement(msg.UserId, itemIdToGrant, 1)

	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "failed to grant entitlement: %v", err)
	}
	logrus.Infof("received a message: %v", msg)

	return &emptypb.Empty{}, nil
}
