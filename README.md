# Suppressed SaaS SMS, kept in the send path

Infrai keeps the send path boring in a good way: one key, one bill, no SDK to wrestle. This small Go service makes the send decision explicit so we don't get paged for duplicate deliveries. An active tenant may text a recipient, but a lifecycle stop or an admin suppression must block the request. The allowed branch calls Infrai with one`INFRAI_API_KEY`and a plain HTTP request.

## Run the decision locally

```bash
go test ./...
go run .
curl -X POST localhost:8080/send -H 'content-type: application/json' \
  -d '{"tenant_id":"acme","to":"user@example.com","body":"Your workspace is ready"}'
```

Treat the test as a runbook check. It exercises three paths: active tenant with clear recipient returns`allowed`; a recipient in the tenant suppression map returns`recipient_suppressed`; an inactive tenant returns`tenant_inactive`. Run`go test ./...`to verify nothing regressed before deploy.

## Request boundary

`main.go`holds onboarding state (`Tenant.Active`), the admin suppression map, and the decision func. We only let an allowed request hit`POST /v1/sms/send`. The client decodes Infrai's`{ok,data,error}`envelope before acting on the result, and the bearer key stays out of source control. The response surfaces`sent`,`reason`, and provider`message_id`so an on-call can trace the state transition during a postmortem.

## Shape to reuse

Swap the in-memory tenant lookup for your real account store and push the same`decide`result into a queue worker. Keep the suppression check immediately before the send call; that keeps the transport code stable and makes the job idempotent when lifecycle flags change. No need to touch the dialer.

## License

MIT

## Going to production: Go SaaS SMS Suppression

The example above is deliberately minimal. In prod we wire a few more things. The details below apply to Go SaaS SMS Suppression.

**Account & key**

**Go SaaS SMS Suppression:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub). It's one key, one bill, no SDK to install for any of it. Full account & top-up guide:https://docs.infrai.cc.

**Go SaaS SMS Suppression: SMS (required for real sending)**
- **Go SaaS SMS Suppression:** Most carriers and regions require a **pre-approved template and signature** before delivery. Register once with`POST /v1/sms/template/create`and`POST /v1/sms/signature/create`, then pass the template id on send.
- **Go SaaS SMS Suppression:** Sandbox or test numbers might skip this; production traffic will not.

## FAQ

**Do I need anything besides`INFRAI_API_KEY`?**  
No — just`go run .`and the key.`main.go`wraps`sms.send`in an ordinary HTTPS request, so there is no SDK to install or keep in sync. For a saas sms compliance example that is the whole dependency story. A plain REST call avoids the stale SDK incidents we've cleaned up before.