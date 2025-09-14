import * as sts from "@aws-sdk/client-sts"

var __callerIdentity: Promise<sts.GetCallerIdentityResponse> | undefined
export async function lookupAccountId(): Promise<string> {
    return callerIdentity().then(id => id.Account!!)
}

export async function lookupUserId(): Promise<string> {
    return callerIdentity().then(id => id.UserId!!)
}

async function callerIdentity(): Promise<sts.GetCallerIdentityResponse> {
    if (__callerIdentity === undefined) {
        const client = new sts.STSClient()
        __callerIdentity = client.send(new sts.GetCallerIdentityCommand({}))
    }
    return __callerIdentity
}