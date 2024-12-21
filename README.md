# Simple OIDC

The codebase is golang (code/lambda) and typescript (infrastructure/deploy).

# Codebase considerations
Dependencies can either be all-in or all-out - and javascript and golang both have very differing philosophies.
At the moment, dependencies will be all-out (node_modules and golang vendoring)

# Aims
1) To provide a simple all in one oidc server that "does enough" to use JWT's as the auth mechanism in an application
2) To provide a application that can easily be swapped out to different storage mechanisms

# Deployment (one touch)
todo - either golang or typescript... or bash...
single script

# Deployment (summary)
1) Ensure your environment is set up
    * AWS Credentials (for default deployment)
    * URL Mount point details (to hook up to an existing aws Route53 zone)
2) Build the binary: 
    - `go generate cmd/gen/gen.go`
    - `go build cmd/simple-oidc/bootstrap.go`
3) Deploy infrastructure: Deployment is run via `npm run cdk deploy simple-oidc`

# Usage or deployed service
1) start off with an authorize reqeust!
