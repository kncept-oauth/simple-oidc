#!/usr/bin/env node
import 'source-map-support/register'
import * as cdk from 'aws-cdk-lib'
import { env } from 'process'

// javascript workaround - need access to async code in the top level function
async function init() {
  const hostedUrl = env.HOSTED_URL

  if (!hostedUrl) {
    console.log('Please specify a HOSTED_URL eg: simple-oidc.kncept.com')
  }

  const app = new cdk.App()
  const oidcStack = new cdk.Stack(app, 'simple-oidc')

}

// can't "await init", but it works from the top level stack
init()
