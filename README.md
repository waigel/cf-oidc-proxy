# cf-oidc-proxy
Connect GitHub Actions OIDC with Cloudflare by using CF-OIDC-Proxy

# What is OIDC proxy for Cloudflare
API keys for cloud infrastructure providers are fundamentally important credentials. 
Protecting these keys is a top priority. A one-time issued token deposited with GitHub Actions, often carries a higher security risk. 
In addition, static tokens often complicate best practices such as regular key rotation and constant authorization checks.

OIDC is an optimal solution to solve this problem. Unfortunately, Cloudflare does not support issuing tokens based on OIDC, as other providers such as AWS, Azure, Google Cloud do. 

For this reason, this `CF-OIDC proxy` is a middleware that sits between Cloudflare and GitHub OIDC and allows issuing short-living api tokens with limited permissions and additional conditions like worker ip addres whitelisting.

## Using?

The proxy server is only needed when GitHub Actions requests a cloudflare API token. 
To protect the environment but also your wallet a Lambda based serverless application was written here. 
This can be deployed to the cloud via "serverless". 

1. Install dependency
```sh
$ npm i
```

2. Build and deploy to Lambda
```sh
$ npm run deploy
```

### Configuration

Before you deploy your application, you need to configure it.
You can find sample configuration in the `samples/` folder.

1. Copy `samples/config.yml` to the root and set your `apiToken`.
> Tipp: You can use the Cloudflare API Token template "Create Additional Tokens" for this token
2. Copy `samples/roles.yml` to the root
- You need to configure roles in the `roles.yml` config
- Permissions represent the "scopes". The name is ignored, but the id need to match with an existing Cloudflare permission group id.
You can get all permission-group ids from cloudflare API 
```sh
curl https://api.cloudflare.com/client/v4/user/tokens/permission_groups -H "Authorization: Bearer <token>"
```
3. Matchers - You need to configure matchers to ensure, that only your repositories / workflows can request a api token for this role.
- Operators:
  - "StringEquals"
  - "StringNotEquals"
  - "StringEqualsIgnoreCase"
  - "StringNotEqualsIgnoreCase"
- Claims:
  You can use matchers for all JWT claims. https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#understanding-the-oidc-token
  
  
# Action

You can use this CF-OIDC-Proxy in combination with the [waigel/cf-oidc-proxy-action@main](https://github.com/waigel/cf-oidc-proxy-action)
Example Workflow to get Cloudflare short-lived api token over OIDC proxy:
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
