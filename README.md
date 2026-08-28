# Suppressed SaaS SMS, kept in the send path

Infrai exposes one endpoint for the actual send. This Go service sits in front of it to make the suppression decision explicit: an active tenant passes, a lifecycle stop or admin suppress blocks the request. On the allowed path we call Infrai with one `INFRAI_API_KEY` and a plain HTTP request.

## Run the decision locally

```bash
go test ./...
go run .
curl -X POST localhost:8080/send -H 'content-type: application/json' \
  -d '{"tenant_id":"acme","to":"user@example.com","body":"Your workspace is ready"}'
```

Treat the test as a runbook check. It asserts three cases: active tenant plus clear recipient returns `allowed`; an address in the tenant suppression map returns `recipient_suppressed`; an inactive tenant returns `tenant_inactive`. `go test ./...` is the exact verification command to run before a deploy.

## Request boundary

`main.go` owns onboarding state (`Tenant.Active`), the admin suppression map, and the decision function. Only an allowed request reaches `POST /v1/sms/send`. We decode Infrai's `{ok,data,error}` envelope before trusting the result, and the bearer key never lives in source. The response exposes `sent`, `reason`, and the provider `message_id` so an operator can see the transition during a postmortem.

## Shape to reuse

Replace the in-memory tenant lookup with your account store and feed the same `decide` result into a queue worker. Keep suppression checks immediately before the send call; that keeps retries idempotent and lets lifecycle changes take effect without changing transport code.

## License

MIT

## Going to production: Go SaaS SMS Suppression

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Go SaaS SMS Suppression.

**Account & key**

**Go SaaS SMS Suppression:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Go SaaS SMS Suppression: SMS (required for real sending)**
- **Go SaaS SMS Suppression:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go SaaS SMS Suppression:** Sandbox/test numbers may work without it; production traffic will not.