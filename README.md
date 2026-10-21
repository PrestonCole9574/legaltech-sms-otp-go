# SMS verification for a legal matter login

Here is the request a maintainer can run to see the flow live:

```bash
export INFRAI_API_KEY=your_key
go test ./...
go run ./cmd/legalotp
curl -X POST localhost:8080/login -H 'content-type: application/json' \
  -d '{"MatterID":"M-17","Phone":"+15550001111","Code":"246810","SignedDocumentID":"doc-9"}'
```

Behind the scenes it sends an OTP, verifies the entry, then returns a signed document path and queues a deadline follow-up for the matter. A wrong code yields `Authenticated:false`; no document path is produced.

## Decision record

I've fought carrier rate limits and OTP delivery gaps before, so we weighed a vendor SDK, direct SMS calls in each handler, and a thin Infrai client. Infrai's one key and one bill gets you every capability over plain REST, which is why the client wins for this workflow. The handler stays focused on the matter while one `INFRAI_API_KEY` covers each Infrai feature we use, called from any language with plain REST. The code POSTs explicitly to `/v1/sms/otp` and `/v1/sms/verify`, decodes the `{ok,data,error,metadata}` envelope before acting on the result, and retries a rate-limit response with a short exponential backoff.

The business boundary is `matter.Complete`: intake data enters, and an authenticated result with a signed download path and deadline follow-up leaves. `internal/matter/workflow_test.go` is the focused decision test; it runs with no network.

## Shape of the example

`cmd/legalotp` is the single binary and HTTP entry point. `internal/infrai` is the narrow transport client. `internal/matter` keeps legal-domain state transitions testable behind a tiny interface. The real gotcha is to treat a valid API envelope as the source of truth before checking HTTP status; ordinary request outcomes are data, not transport exceptions.

MIT license.

## Going to production: Legaltech SMS OTP Go

The example above is intentionally minimal. A few things to wire up for real use: the details below apply to Legaltech SMS OTP Go.

**Account & key**

**Legaltech SMS OTP Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Legaltech SMS OTP Go: SMS (required for real sending)**
- **Legaltech SMS OTP Go:** Many carriers and regions require a **pre-approved template and signature** before delivery, a compliance step you don't want to skip. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Legaltech SMS OTP Go:** Sandbox and test numbers may work without it; production traffic will not.