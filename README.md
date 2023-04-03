# cf-oidc-proxy
Connect GitHub Actions OIDC with Cloudflare using CF-OIDC-Proxy.

# Introduction
API keys for cloud infrastructure providers are important credentials that must be protected. A one-time issued token deposited with GitHub Actions carries a higher security risk. Additionally, static tokens complicate best practices such as regular key rotation and constant authorization checks.

OIDC is an optimal solution to this problem. However, Cloudflare does not support issuing tokens based on OIDC, unlike other providers such as AWS, Azure, and Google Cloud.

To solve this problem, the CF-OIDC-Proxy acts as a middleware between Cloudflare and GitHub OIDC, allowing the issuance of short-living API tokens with limited permissions and additional conditions like IP address whitelisting for workers.

## Using?

The proxy server is only required when GitHub Actions requests a Cloudflare API token. To protect the environment and your wallet, a Lambda-based serverless application was written that can be deployed to the cloud via "serverless."

1. Install dependencies:
```sh
$ npm i
```

2. Build and deploy to Lambda:
```sh
$ npm run deploy
```

### Configuration

Before deploying your application, you need to configure it. A sample configuration can be found in the samples/ folder.

1. Copy `samples/config.yml` to the root and set your apiToken. You can use the Cloudflare API Token template "Create Additional Tokens" for this step.
2. Copy `samples/roles.yml` to the root. You need to configure roles in the `roles.yml` config. Permissions represent the "scopes," and the name is ignored, but the ID needs to match an existing Cloudflare permission group ID. You can get all permission group IDs from Cloudflare API using:
```sh
curl https://api.cloudflare.com/client/v4/user/tokens/permission_groups -H "Authorization: Bearer <token>"
```
3. Matchers: You need to configure matchers to ensure that only your repositories/workflows can request an API token for this role. Operators include:
- "StringEquals"
- "StringNotEquals"
- "StringEqualsIgnoreCase"
- "StringNotEqualsIgnoreCase"
Claims: You can use matchers for all JWT claims. See https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token for more information.
  
# Action

You can use the CF-OIDC-Proxy with waigel/cf-oidc-proxy-action@main to get a Cloudflare short-lived API token over the OIDC proxy.

Example workflow:

```yaml
name: Cloudflare OIDC Test
on:
  workflow_dispatch:
  
permissions:
  id-token: write
jobs:
  cloudflare:
    runs-on: ubuntu-latest
    steps:
      - uses: waigel/cf-oidc-proxy-action@main
        id: cloudflare
        with:
          proxy-url: https://<lambda-id>.execute-api.eu-central-1.amazonaws.com
          role-to-assume: dns
      - name: Verify API token is valid
        run: |
          curl "https://api.cloudflare.com/client/v4/user/tokens/verify" \
          -H  "Authorization: Bearer ${{ steps.cloudflare.outputs.api_token }}" \
          | grep -o '"message":"[^"]*"' | sed 's/"message":"\(.*\)"/\1/
```
---
This project is licensed under the MIT License.
