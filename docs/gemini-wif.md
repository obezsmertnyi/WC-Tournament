# Calling Gemini (Vertex AI) from the WC-Tournament backend

How to add Gemini calls to the Go backend **without any API key or JSON service-account
key**. Auth is Workload Identity Federation (WIF): a Teleport `tbot` on the host keeps a
short-lived credential fresh, the Google SDK picks it up via ADC and auto-refreshes. There
is **nothing to rotate and no 1-hour token to refresh** in app code.

Project: `<GCP_PROJECT>` · Region: `us-central1` · SA: `gemini-agent@<GCP_PROJECT>.iam.gserviceaccount.com` (has `roles/aiplatform.user`).

## What the host must provide (infra, not app code)

The deploy target must already have (see the infra runbook, below):

1. `tbot` (systemd) writing a JWT-SVID to `/opt/workload-identity/jwt_svid` and keeping it fresh.
2. A WIF **credential-config** JSON (external_account) that points ADC at that JWT and
   impersonates the SA. Generate it **once** per host:
   ```bash
   gcloud iam workload-identity-pools create-cred-config \
     projects/<GCP_PROJECT_NUMBER>/locations/global/workloadIdentityPools/<WIF_POOL>/providers/<WIF_PROVIDER> \
     --service-account=gemini-agent@<GCP_PROJECT>.iam.gserviceaccount.com \
     --credential-source-file=/opt/workload-identity/jwt_svid \
     --credential-source-type=text \
     --output-file=/etc/gemini/cred.json
   ```
   This file contains **no secret** — only pointers; the secret is the JWT tbot maintains.

## Environment (set for the backend process)

```bash
GOOGLE_APPLICATION_CREDENTIALS=/etc/gemini/cred.json   # ADC reads the WIF cred-config
GOOGLE_GENAI_USE_VERTEXAI=true
GOOGLE_CLOUD_PROJECT=<GCP_PROJECT>
GOOGLE_CLOUD_LOCATION=us-central1
```

## Go code (`google.golang.org/genai`)

```bash
go get google.golang.org/genai
```

```go
package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// NewClient builds a Vertex-backed Gemini client. Credentials come from ADC
// (GOOGLE_APPLICATION_CREDENTIALS -> WIF cred-config -> tbot JWT). The access
// token is minted and auto-refreshed by the SDK; nothing to refresh here.
func NewClient(ctx context.Context) (*genai.Client, error) {
	return genai.NewClient(ctx, &genai.ClientConfig{
		Project:  "<GCP_PROJECT>",
		Location: "us-central1",
		Backend:  genai.BackendVertexAI,
	})
}

// Ask sends one prompt and returns the model's text.
func Ask(ctx context.Context, c *genai.Client, model, prompt string) (string, error) {
	resp, err := c.Models.GenerateContent(ctx, model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	return resp.Text(), nil
}
```

Usage:

```go
c, err := gemini.NewClient(ctx)      // reuse across requests; it is concurrency-safe
// ...
out, err := gemini.Ask(ctx, c, "gemini-2.5-flash", "Summarise this fixture in one line: ...")
```

## Models

- `gemini-2.5-flash` — default; fast, cheap, good quality.
- `gemini-2.5-flash-lite` — cheapest; short/simple tasks.
- `gemini-2.5-pro` — hardest reasoning.
- `gemini-embedding-001` — embeddings (`c.Models.EmbedContent`).

Model choice is per-call; the SA has `aiplatform.user`, so any Vertex model in the project works.

## Rules

- **Do not** put an API key, a JSON SA key, or a static bearer token in code, env, or config.
- **Do not** cache the access token yourself — the SDK refreshes it. Caching a token
  reintroduces the exact 1-hour-expiry bug this setup removes.
- **Do** reuse one `*genai.Client` process-wide.
- Local dev without tbot: `gcloud auth application-default login` also satisfies ADC, so the
  same code runs unchanged on a laptop.

## Debugging the credential chain (no Go)

To confirm the host's WIF chain end-to-end from a shell (JWT → STS → impersonate → Vertex):

```bash
sudo gemini-check gemini-2.5-flash "Reply OK"
```

(`gemini-check` is the pure-curl verifier installed on WIF hosts.) If that returns text, the
Go client will too. Full trust-chain explanation: `gcp-gemini-int/docs/vertex-wif-as-built.md` §2.1.
