package server

import (
	"context"
	pb "extend-event-listener/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
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

type OauthHandler struct {
	pb.UnimplementedIAMServiceOAuthEventsServer
	entitlement platform.EntitlementService
}

func NewOauthHandler(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
) *OauthHandler {
	return &OauthHandler{
		entitlement: platform.EntitlementService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
	}
}

func (o *OauthHandler) grantEntitlement(userID string, itemID string, count int32) error {
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

func (o *OauthHandler) PublishOauthTokenGeneratedMessageToOauthTokenChannel(ctx context.Context, msg *pb.OauthTokenGeneratedMessage) (*emptypb.Empty, error) {
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
