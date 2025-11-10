// Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
)

// grantEntitlement is a shared function to grant entitlements to users
func grantEntitlement(fulfillmentService platform.FulfillmentService, namespace string, userID string, itemID string) error {
	quantity := int32(1)
	fulfillmentResponse, err := fulfillmentService.FulfillItemShort(&fulfillment.FulfillItemParams{
		Namespace: namespace,
		UserID:    userID,
		Body: &platformclientmodels.FulfillmentRequest{
			ItemID:   itemID,
			Quantity: &quantity,
			Source:   platformclientmodels.EntitlementGrantSourceREWARD,
		},
	})

	if err != nil {
		return err
	}

	if fulfillmentResponse == nil || fulfillmentResponse.EntitlementSummaries == nil || len(fulfillmentResponse.EntitlementSummaries) <= 0 {
		return fmt.Errorf("could not grant item to user")
	}

	return nil
}
