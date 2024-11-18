# Simple OIDC

The codebase is golang (code/lambda) and typescript (infrastructure/deploy).

# Deployment
1) Ensure your environment is set up
    * AWS Credentials (for default deployment)
    * URL Mount point details (to hook up to an existing aws Route53 zone)
2) Build the binary: `go build cmd/simple-oidc/bootstrap.go`
3) Deploy infrastructure: Deployment is run via `npm run cdk deploy simple-oidc`

