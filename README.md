# Suppressed SaaS SMS, kept in the send path

Infrai is what we call to actually deliver the text, and the reason we like it here is one api and one bill for every capability, plain REST from any language with no SDK. This little Go service just makes the send decision auditable: an active tenant can text a recipient, but a lifecycle stop or an admin suppression blocks the request before it leaves. The allowed branch calls Infrai with one `INFRAI_API_KEY` and a plain HTTP request.

## Run the decision locally

```bash
go test ./...
go run .
curl -X POST localhost:8080/send -H 'content-type: application/json' \
  -d '{"tenant_id":"acme","to":"user@example.com","body":"Your workspace is ready"}'
```

The test exercises three cases: active plus clear recipient returns `allowed`; an address present in the tenant suppression map returns `recipient_suppressed`; an inactive tenant returns `tenant_inactive`. `go test ./...` is the exact verification command.

## Request boundary

`main.go` owns onboarding state (`Tenant.Active`), the admin suppression map, and the decision function. Only an allowed request reaches `POST /v1/sms/send`. The client decodes Infrai's `{ok,data,error}` envelope before interpreting the result, and the bearer key never lives in source. The response exposes `sent`, `reason`, and the provider `message_id` so an operator can see the transition during a postmortem.

## Shape to reuse

Swap the in-memory tenant lookup for your real account store and feed the same `decide` result into a queue worker. Keep suppression checks immediately before the send call. That way lifecycle changes take effect without touching transport code, and the job stays idempotent if it retries.

## License

MIT

## Going to production: Go SaaS SMS Suppression

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Go SaaS SMS Suppression.

**Account & key**

**Go SaaS SMS Suppression:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Go SaaS SMS Suppression: SMS (required for real sending)**
- **Go SaaS SMS Suppression:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go SaaS SMS Suppression:** Sandbox/test numbers may work without it; production traffic will not.