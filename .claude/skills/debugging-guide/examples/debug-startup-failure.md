# Example: Service Exits Immediately at Startup

This annotated example shows a debugging session where the Extend Event Handler process
exits right after launch with no gRPC server ever starting.

---

## The Report

A developer sends this message:

> *"I run `go run main.go` and the process just quits. Nothing happens, no server, no
> useful error. I set up `.env` from the template."*

---

## Step 1 — Read the exit output carefully

Ask the developer to share the full terminal output. They share:

```
2026-03-10T08:45:01Z INFO  starting extend-event-handler
ITEM_ID_TO_GRANT env is not set or empty
exit status 1
```

**What this tells us:**
- The service reached the env validation block in `main.go`.
- A required app-specific env var (`ITEM_ID_TO_GRANT`) is missing from the environment.
- The process called `os.Exit(1)` before ever starting the gRPC server.

Note: the standard `AB_*` variables are present (the IAM login error would appear instead
if they were missing). This is a domain-specific variable the template may not have prompted for.

---

## Step 2 — Find the validation in main.go

Read `main.go` to confirm what the service expects:

```go
itemIdToGrant := os.Getenv("ITEM_ID_TO_GRANT")
if itemIdToGrant == "" {
    log.Fatal("ITEM_ID_TO_GRANT env is not set or empty")
}
```

**Problem identified:** The `.env.template` lists `ITEM_ID_TO_GRANT=` with a blank value,
and the developer copied it without filling it in.

---

## Step 3 — Check the .env file

```bash
grep ITEM_ID_TO_GRANT .env
# Output: ITEM_ID_TO_GRANT=
```

Confirmed — the value is empty.

---

## Step 4 — The fix

Open `.env` and set the real item ID from the AccelByte platform:

```
ITEM_ID_TO_GRANT=a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

To find the correct item ID: go to your AGS environment → Publisher Portal → Game
Management → Items, and copy the ID of the item you want to grant on login.

> **Important:** Every app-specific required variable like this must be filled in before
> the service can start. Check `.env.template` for any variable with a blank default and
> confirm you have a real value for each one.

---

## Step 5 — Verify

```bash
go run main.go
# Expected output (no exit):
# 2026-03-10T08:50:01Z INFO  starting extend-event-handler
# 2026-03-10T08:50:01Z INFO  IAM login successful
# 2026-03-10T08:50:01Z INFO  gRPC server listening on :6565
```

```bash
# Confirm gRPC server is up
ss -tlnp | grep 6565
# Expected: LISTEN ... :6565
```

---

## Key takeaways from this session

1. **First line of the fatal log is the diagnosis.** The process printed the exact missing
   variable name before exiting — no breakpoint needed.
2. **App-specific vars are not always in README.** Always read `main.go` startup code to
   find every `os.Getenv` call that exits on an empty value. These are hard dependencies.
3. **`.env.template` blank values are intentional placeholders** — they will not work as-is.
   Fill every blank value before running the service for the first time.
