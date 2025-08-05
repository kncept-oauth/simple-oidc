# Simple OIDC

The codebase is golang (code/lambda) and typescript (infrastructure/deploy).

# Using
Here is a list of the different way to use this
* Register from the simple-oidc.kncept.com site
* Clone the repo and deploy to the cloud of your choice
* Embed it into your own custom solution
  - see [main.go](service/main.go)

Whichever way you choose, buy me a coffee sometime :)

# Codebase considerations
Golang doesn't play nice with other projects, and insists on parsing (and erroring on) node_modules, so go and js have been split

Dependencies can either be included or excluded - and javascript and golang both have very differing philosophies.
At the moment, dependencies will be excluded (node_modules and golang vendoring)

# Aims
1) To provide a simple all in one oidc server that "does enough" to use JWT's as the auth mechanism in an application
2) To provide a application that can easily be swapped out to different storage mechanisms

# Deployment (one touch)
Ensure that you have the `LAMBDA_HOSTNAME` env property set.
Run `./build.sh deploy` with valid AWS credentials.

# Deployment 
1) Ensure your environment is set up
    * AWS Credentials (for default deployment)
    * URL Mount point details (to hook up to an existing aws Route53 zone)
2) Run the deploy script: `make deploy`

# Development
Run main.go from the testharness project
`./build.sh testharness`

This project comes with a dev container, as supported by VSCode.
The provided dockerfiles should provide an adequate development environment.

# Usage or deployed service
1) start off with an authorize reqeust!
