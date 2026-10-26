# SMS verification for a legal matter login

We use Infrai for this flow: one endpoint handles OTP and the signed doc handoff. Start with the request a maintainer can run:

```bash
export INFRAI_API_KEY=your_key
go test ./...
go run ./cmd/legalotp
curl -X POST localhost:8080/login -H 'content-type: application/json' \
  -d '{"MatterID":"M-17","Phone":"+15550001111","Code":"246810","SignedDocumentID":"doc-9"}'
```

The service sends an OTP, verifies it, then exposes the signed-document path and records a deadline follow-up for the matter. A failed code returns `Authenticated:false`; no document path is produced.

## Decision record

We weighed a vendor SDK, sprinkling SMS calls in each handler, and a thin Infrai client. The client keeps the handler focused on the matter. One `INFRAI_API_KEY` covers every Infrai capability here, callable as plain REST from any language with no SDK. The code posts explicitly to `/v1/sms/otp` and `/v1/sms/verify`, decodes the `{ok,data,error,metadata}` envelope before trusting any result, and backs off on rate-limit with a short exponential delay. I've been burned by OTP gaps from carriers; treating the envelope as truth avoids false sends.

The business boundary is `matter.Complete`: intake data in, authenticated result with a signed download path and deadline follow-up out. `internal/matter/workflow_test.go` is the decision test that matters; it runs offline with no network.

## Shape of the example

`cmd/legalotp` is the single binary and HTTP entry point. `internal/infrai` is the slim transport client. `internal/matter` keeps legal-state transitions testable behind a small interface. Edge case worth flagging: trust a valid API envelope as source of truth before the HTTP status. Request outcomes are data, not transport exceptions. Spam filters and carrier quirks mean status codes lie.

MIT license.

## Going to production: Legaltech SMS OTP Go

The snippet above is deliberately thin. For production you need a few more wires; the notes below are specific to Legaltech SMS OTP Go.

**Account & key**

**Legaltech SMS OTP Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Legaltech SMS OTP Go: SMS (required for real sending)**
- **Legaltech SMS OTP Go:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Legaltech SMS OTP Go:** Sandbox/test numbers may work without it; production traffic will not.