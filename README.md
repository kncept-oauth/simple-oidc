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

Dependencies can either be all-in or all-out - and javascript and golang both have very differing philosophies.
At the moment, dependencies will be all-out (node_modules and golang vendoring)
I expect that this will switch to all-in before long


# Aims
1) To provide a simple all in one oidc server that "does enough" to use JWT's as the auth mechanism in an application
2) To provide a application that can easily be swapped out to different storage mechanisms

# Deployment (one touch)
todo - either golang or typescript... or bash...
single script

# Deployment 
1) Ensure your environment is set up
    * AWS Credentials (for default deployment)
    * URL Mount point details (to hook up to an existing aws Route53 zone)
2) Run the deploy script: `make deploy`

# Development
Run main.go from the testharness project
`cd testharness && go run main.go`

# Usage or deployed service
1) start off with an authorize reqeust!
