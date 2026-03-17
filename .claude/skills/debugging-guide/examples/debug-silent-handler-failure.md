# Example: Handler Called but Downstream Action Not Performed

This annotated example shows a debugging session where events are being received and logged,
but the downstream action (granting an entitlement) silently does nothing.

---

## The Report

A developer sends this message:

> *"The service is running and I can see in the logs that `OnMessage` is being called when
> a user logs in, but the item is never granted. No errors in the logs either."*

---

## Step 1 — Read the logs at debug level

The developer initially has `LOG_LEVEL=info`. Ask them to restart with `LOG_LEVEL=debug`
and share the output during a login event.

They share:

```json
{"time":"2026-03-10T11:20:01Z","level":"INFO","msg":"received an event","traceId":"abc123","handler":"LoginHandler.OnMessage","userId":"user-xyz"}
{"time":"2026-03-10T11:20:01Z","level":"INFO","msg":"finished call","grpc.service":"UserAuthenticationUserLoggedInService","grpc.method":"OnMessage","grpc.code":"OK"}
```

**What this tells us:**
- The event was received and the handler returned `OK`.
- There is no ERROR log — but also no log confirming the item was granted.
- The handler completed without error, yet the action didn't happen. This points to a
  silent failure or a missing log inside `grantEntitlement`.

---

## Step 2 — Read the business logic

Read the handler's `OnMessage` and the downstream function it calls. In this repo:

```go
// pkg/service/loginHandler.go
func (o *LoginHandler) OnMessage(ctx context.Context, msg *pb.UserLoggedIn) (*emptypb.Empty, error) {
    scope := common.GetScopeFromContext(ctx, "LoginHandler.OnMessage")
    defer scope.Finish()
    scope.Log.Info("received an event", "userId", msg.UserId)
    err := grantEntitlement(o.fulfillment, o.namespace, msg.UserId, itemIdToGrant)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to grant entitlement: %v", err)
    }
    return &emptypb.Empty{}, nil
}
```

```go
// pkg/service/entitlement.go
func grantEntitlement(...) error {
    fulfillmentResponse, err := fulfillmentService.FulfillItemShort(...)
    if err != nil {
        return err
    }
    if fulfillmentResponse == nil {
        return errors.New("nil fulfillment response")
    }
    if len(fulfillmentResponse.GetPayload().EntitlementSummaries) == 0 {
        // ← No return error here — falls through and returns nil
        slog.Warn("entitlement summaries empty")
    }
    return nil
}
```

**Problem identified:** When the Platform SDK returns an empty `EntitlementSummaries` list,
the function logs a warning but returns `nil` (success). The caller sees no error, and
`OnMessage` returns `OK` — but no entitlement was actually created. This happens when the
`ITEM_ID_TO_GRANT` points to an item that exists in the catalogue but is not configured for
fulfilment in the target namespace.

---

## Step 3 — Confirm with a breakpoint

Set a breakpoint at the `len(fulfillmentResponse.GetPayload().EntitlementSummaries) == 0`
check in `entitlement.go`. When the breakpoint hits:

1. Inspect `fulfillmentResponse` in the Variables panel — check `Payload.EntitlementSummaries`.
2. Inspect `itemIdToGrant` — verify it matches the item ID configured in your AGS environment.

---

## Step 4 — Root cause and fix

The item ID is valid but the item is not set up for the `REWARD` fulfilment source in
this namespace. There are two ways to fix this:

**Option A — Fix the item configuration in AGS (preferred):**
Go to Publisher Portal → Game Management → Items → find the item → ensure it is published
and the fulfilment source `REWARD` is enabled for the target namespace.

**Option B — Surface the silent failure as an error** (improves observability regardless):

```go
// pkg/service/entitlement.go — before
if len(fulfillmentResponse.GetPayload().EntitlementSummaries) == 0 {
    slog.Warn("entitlement summaries empty")
}
return nil

// After
if len(fulfillmentResponse.GetPayload().EntitlementSummaries) == 0 {
    return errors.New("entitlement granted but EntitlementSummaries is empty — item may not be configured for REWARD fulfilment")
}
return nil
```

With Option B the handler returns `codes.Internal`, the log shows `"level":"ERROR"`, and the
silent failure becomes visible immediately.

---

## Step 5 — Verify

After fixing the item configuration or surfacing the error:

```bash
# Trigger a login event by calling grpcurl manually
grpcurl -plaintext -d '{"userId": "test-user", "namespace": "mygame"}' \
  localhost:6565 \
  accelbyte.asyncapi.iam.account.v1.UserAuthenticationUserLoggedInService/OnMessage
# Expected: {}  (empty response = success)
```

Check the AGS console or use the Platform API to confirm an entitlement now appears for
`test-user` under the target item ID.

---

## Key takeaways from this session

1. **`codes.OK` does not mean the business action succeeded.** A handler can return `OK`
   even if the downstream SDK call produced no meaningful result. Always log a success
   confirmation after the action, not just after the handler exits.
2. **Warn-and-continue is a debugging trap.** A warning that doesn't surface as an error
   makes silent failures invisible. Treat unexpected empty responses as errors.
3. **`LOG_LEVEL=debug` is essential.** At `info` level, only the entry and exit of each
   handler call are logged. Debug level reveals the SDK call results and intermediate state.
4. **Confirm the item configuration, not just the code.** Half of "item not granted" issues
   are AGS configuration problems, not code bugs. Check the item's fulfilment settings in
   the Publisher Portal before assuming the code is wrong.
