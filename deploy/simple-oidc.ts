#!/usr/bin/env node
import 'source-map-support/register'
import * as cdk from 'aws-cdk-lib'
import { env } from 'process'
import * as iam from 'aws-cdk-lib/aws-iam'
import * as logs from 'aws-cdk-lib/aws-logs'
import * as lambda from 'aws-cdk-lib/aws-lambda'
import * as apigateway from 'aws-cdk-lib/aws-apigateway'
import { matchingHostedZone } from './lib/domain-tools'

// javascript workaround - need access to async code in the top level function
async function init() {
  const lambdaHostname = env.LAMBDA_HOSTNAME || 'simple-oidc.kncept.com'
  // if (!hostedUrl) {
  //   console.log('Please specify a HOSTED_URL eg: simple-oidc.kncept.com')
  // }

  const app = new cdk.App()
  const stack = new cdk.Stack(app, 'simple-oidc')


  const role = new iam.Role(stack, `simple-oidc-role`, {
    assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
    roleName: `simple-oidc-role`,
    managedPolicies: [
      // iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaVPCAccessExecutionRole"),
      iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
    ],
  })

  const logGroup = new logs.LogGroup(stack, `logs`, {
    logGroupName: `/logs/simple-oidc`,
    retention: logs.RetentionDays.THREE_MONTHS,
    removalPolicy: cdk.RemovalPolicy.DESTROY,
  })

  logGroup.grantWrite(role)

  const fn = new lambda.Function(stack, `simple-oidc-fn `, {
    runtime: lambda.Runtime.PROVIDED_AL2023,
    functionName: `simple-oidc-fn`,
    code: lambda.Code.fromAsset('../service', {exclude: ['**', '!bootstrap']}),
    handler: 'bootstrap',
    role,
    logGroup,
    environment: {
      // no dot . or dash - allowed
    },
  })

      const handlerIntegration = new apigateway.LambdaIntegration(fn, {
      allowTestInvoke: false,
    })

 const restApi = new apigateway.RestApi(stack, `simple-oidc-api`, {
      restApiName: 'Simple OIDC',
      description: 'Kncept Simple OIDC and Oauth2 Server',
      endpointTypes: [apigateway.EndpointType.REGIONAL],
      minCompressionSize: cdk.Size.bytes(0),
      defaultIntegration: handlerIntegration,
    })

    // restApi.root.addProxy({
    //   defaultIntegration: handlerIntegration
    // })
    // restApi.root.addMethod('GET', handlerIntegration, {})


    const hostedZoneInfo = await matchingHostedZone(lambdaHostname)

    // find the LONGEST matching hostnamd possible
    
}

// can't "await init", but it works from the top level stack
init()
