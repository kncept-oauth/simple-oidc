import * as jsonTables from './tables.json'

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
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb'


import { lookupAccountId } from './lib/account-tools'

// the FIRST region will be where the data stack also ends up
const regions = process.env.REGIONS || 'ap-southeast-2'
const allRegions = regions.split(",")
for (let i = 0; i < allRegions.length; i++) {
  allRegions[i] = allRegions[i].trim()
}
// javascript workaround - need access to async code in the top level function
async function init() {
  const lambdaHostname = env.LAMBDA_HOSTNAME || 'simple-oidc.kncept.com'
  const name = 'simple-oidc'
  const accountId = await lookupAccountId()
  const app = new cdk.App()

  const dataStack = new cdk.Stack(app, `${name}-data`, {
    env: {
      region: allRegions[0]
    }
  })
  const tables = defineDataStack(dataStack)


  await Promise.all(allRegions.map(async region => {
    const appStack = new cdk.Stack(app, `${name}-app-${region}`, {
      env: {
        region,
      }
    })

    const role = new iam.Role(appStack, `${name}-role`, {
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
      roleName: `${name}-role`,
      managedPolicies: [
        // iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaVPCAccessExecutionRole"),
        iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
      ],
    })

    const logGroup = new logs.LogGroup(appStack, `${name}-logs`, {
      logGroupName: `/logs/${name}`,
      retention: logs.RetentionDays.THREE_MONTHS,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    })

    logGroup.grantWrite(role)

    const fn = new lambda.Function(appStack, `${name}-fn `, {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      functionName: `${name}-fn`,
      code: lambda.Code.fromAsset('../service', { exclude: ['**', '!bootstrap'] }),
      handler: 'bootstrap',
      role,
      logGroup,
      environment: {
        // no dot . or dash - allowed

        'host_name': lambdaHostname,

        'git_hash': process.env.GITHUB_SHA || 'unknown',
        'deploytime': `${new Date()}`,
      },
    })

    const restApi = new apigateway.LambdaRestApi(appStack, `${name}-restapi`, {
      restApiName: 'Simple OIDC',
      description: 'Kncept Simple OIDC and Oauth2 Server',
      endpointTypes: [apigateway.EndpointType.REGIONAL],
      handler: fn,
      minCompressionSize: cdk.Size.bytes(0),
    })

    // will match the longest hostname possible
    const hostedZoneInfo = await matchingHostedZone(lambdaHostname)
    const zone = route53.PublicHostedZone.fromHostedZoneAttributes(
      appStack,
      `${name}-zone`,
      {
        hostedZoneId: hostedZoneInfo.id,
        zoneName: hostedZoneInfo.name,
      },
    )

    const certificate = new certificatemanager.Certificate(
      appStack,
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

    ///sf
    new route53.CnameRecord(appStack, `${name}-api-cname`, {
      zone,
      recordName: lambdaHostname.substring(0, lambdaHostname.length - (hostedZoneInfo.name.length + 1)),
      domainName: apiDomainNameMountPoint.domainNameAliasDomainName,
      region,
    })

    role.addToPolicy(new iam.PolicyStatement({
      sid: 'DdbPermissions',
      actions: [
        'dynamodb:BatchGetItem',
        'dynamodb:BatchWriteItem',
        'dynamodb:DeleteItem',
        'dynamodb:DescribeTable',
        'dynamodb:GetItem',
        'dynamodb:ListTables',
        'dynamodb:PutItem',
        'dynamodb:Query',
        'dynamodb:Scan',
        'dynamodb:UpdateItem',
      ],
      // resources: tables.map(table => table.tableArn),
      resources: tables.map(table => `arn:aws:dynamodb:${region}:${accountId}:table/${table.tableName}`), // arn:aws:dynamodb:region:account-id:table/table-name
      effect: iam.Effect.ALLOW,
    }))
    return appStack
  }))
}

interface TableDetails {
  tableName: string
  table: dynamodb.TableV2
}



function defineDataStack(dataStack: cdk.Stack): Array<TableDetails> {
  return jsonTables.tables.map((ddbTableToCreate: any) => {
    const tableName: string = ddbTableToCreate.tableName
    const partitionKeyName: string = ddbTableToCreate.partitionKeyName
    const sortKeyName: string | undefined = ddbTableToCreate.sortKeyName

    const globalSecondaryIndexes: Array<dynamodb.GlobalSecondaryIndexPropsV2> = []
    if (sortKeyName !== undefined) {
      globalSecondaryIndexes.push({
        indexName: "reverse",
        partitionKey: { name: sortKeyName, type: dynamodb.AttributeType.STRING },
        sortKey: { name: partitionKeyName, type: dynamodb.AttributeType.STRING },
      })
    }

    const replicas: Array<dynamodb.ReplicaTableProps> = []
    for (let i = 1; i < allRegions.length; i++) {
      replicas.push({
        region: allRegions[i],
      })
    }

    return {
      tableName,
      table: new dynamodb.TableV2(
        dataStack,
        `table-${tableName}`,
        {
          tableName,
          partitionKey: { name: partitionKeyName, type: dynamodb.AttributeType.STRING },
          sortKey: sortKeyName === undefined ? undefined : { name: sortKeyName, type: dynamodb.AttributeType.STRING },
          timeToLiveAttribute: "ttl",
          removalPolicy: cdk.RemovalPolicy.DESTROY,
          billing: dynamodb.Billing.onDemand(),
          globalSecondaryIndexes,
          replicas,
        }
      )
    }
  })
}

// can't "await init", but it works from the top level stack
init()
