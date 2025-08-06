#!/usr/bin/env node
import 'source-map-support/register'
import * as cdk from 'aws-cdk-lib'
import { env } from 'process'
import * as iam from 'aws-cdk-lib/aws-iam'
import * as logs from 'aws-cdk-lib/aws-logs'
import * as lambda from 'aws-cdk-lib/aws-lambda'
import * as apigateway from 'aws-cdk-lib/aws-apigateway'
import * as certificatemanager from 'aws-cdk-lib/aws-certificatemanager'
import * as route53 from 'aws-cdk-lib/aws-route53'
import { matchingHostedZone } from './lib/domain-tools'

// javascript workaround - need access to async code in the top level function
async function init() {
  const lambdaHostname = env.LAMBDA_HOSTNAME || 'simple-oidc.kncept.com'
  const name = 'simple-oidc'

  const app = new cdk.App()
  const stack = new cdk.Stack(app, name)


  const role = new iam.Role(stack, `${name}-role`, {
    assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
    roleName: `${name}-role`,
    managedPolicies: [
      // iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaVPCAccessExecutionRole"),
      iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
    ],
  })

  const logGroup = new logs.LogGroup(stack, `${name}-logs`, {
    logGroupName: `/logs/${name}`,
    retention: logs.RetentionDays.THREE_MONTHS,
    removalPolicy: cdk.RemovalPolicy.DESTROY,
  })

  logGroup.grantWrite(role)

  const fn = new lambda.Function(stack, `${name}-fn `, {
    runtime: lambda.Runtime.PROVIDED_AL2023,
    functionName: `${name}-fn`,
    code: lambda.Code.fromAsset('../service', { exclude: ['**', '!bootstrap'] }),
    handler: 'bootstrap',
    role,
    logGroup,
    environment: {
      // no dot . or dash - allowed

      'git_hash': process.env.GITHUB_SHA || 'unknown',
      'deploytime': `${new Date()}`,
    },
  })

  const handlerIntegration = new apigateway.LambdaIntegration(fn, {
    allowTestInvoke: false,
  })

  const restApi = new apigateway.RestApi(stack, `${name}-restapi`, {
    restApiName: 'Simple OIDC',
    description: 'Kncept Simple OIDC and Oauth2 Server',
    endpointTypes: [apigateway.EndpointType.REGIONAL],
    minCompressionSize: cdk.Size.bytes(0),
    // defaultIntegration: handlerIntegration,
  })
  restApi.root.addProxy({
    defaultIntegration: handlerIntegration
  })
  // restApi.root.addMethod('GET', handlerIntegration, {})

  // will match the longest hostname possible
  const hostedZoneInfo = await matchingHostedZone(lambdaHostname)
  const zone = route53.PublicHostedZone.fromHostedZoneAttributes(
    stack,
    `${name}-zone`,
    {
      hostedZoneId: hostedZoneInfo.id,
      zoneName: hostedZoneInfo.name,
    },
  )

  const certificate = new certificatemanager.Certificate(
    stack,
    `${name}-cert`,
    {
      domainName: lambdaHostname,
      validation: certificatemanager.CertificateValidation.fromDns(zone),
    }
  )

  const apiDomainNameMountPoint = restApi.addDomainName(`${name}-rest-dn`, {
    domainName: lambdaHostname,
    certificate,
  })

  new route53.CnameRecord(stack, `${name}-api-cname`, {
    zone,
    recordName: lambdaHostname.substring(0, lambdaHostname.length - (hostedZoneInfo.name.length + 1)),
    domainName: apiDomainNameMountPoint.domainNameAliasDomainName,
  })

}

// can't "await init", but it works from the top level stack
init()
