import { Route53Client, ListHostedZonesCommand, ListHostedZonesRequest } from "@aws-sdk/client-route-53"
import { error } from "console"
// import * as route53 from 'aws-cdk-lib/aws-route53'

// yeah, this isn't super clever.
// but it should work for a bunch of things
const domainNameEndingsToScan = [
    '.com',
    '.com.au',
    '.org.au',
    '.co.nz',
    '.co.uk',
]

export type HostedZoneInfo = {
    id: string,
    name: string,
}

export async function matchingHostedZone(lambdaHostname: string): Promise<HostedZoneInfo> {
    const hostedZones = await listDomainNames()
    hostedZones.sort((a: HostedZoneInfo, b: HostedZoneInfo) : number => {
        return a.name.length - b.name.length
    })
    
    for(let i = 0; i < hostedZones.length; i++) {
        const hostedZone = hostedZones[i]
        console.log(hostedZone)
    }

    throw new Error(`No matching hosted zone for ${lambdaHostname}`)
}

export async function listDomainNames() : Promise<Array<HostedZoneInfo>> {
    const client = new Route53Client({})
    const input: ListHostedZonesRequest = {}
    const command = new ListHostedZonesCommand(input)
    const response = await client.send(command)
    if (!response.HostedZones) return []
    return response.HostedZones.map(hz => {
        let id = hz.Id!
        if (id.startsWith('/hostedzone/')) id = id.substring('/hostedzone/'.length)
        let name = hz.Name!
        if (name.endsWith('.')) name = name.substring(0, name.length - 1)
        return {id, name} as HostedZoneInfo})
}

export function extractHostedZoneFromHostname(fqdn: string) : string {
    if (fqdn === undefined || fqdn === null || fqdn.trim() === '') throw new Error('Bad fqdn: ' + fqdn)
    for(let i = 0; i < domainNameEndingsToScan.length; i++) {
        const knownTldEnding = domainNameEndingsToScan[i]
        if (fqdn.endsWith(knownTldEnding)) {
            const hostname = fqdn.substring(0, fqdn.length - knownTldEnding.length)
            if (hostname.lastIndexOf('.') != -1) {
                return fqdn.substring(hostname.lastIndexOf('.') + 1)
            } else {
                return fqdn
            }
        }
    }
    throw new Error('Unable to determine hosted zone name')
}

export async function matchHostedZoneToFQDN(fqdn: string) : Promise<HostedZoneInfo | undefined> {
    const existingDomainNames = await listDomainNames()
    const tld = extractHostedZoneFromHostname(fqdn)
    for(let i = 0; i < domainNameEndingsToScan.length; i++) {
        let existingDomainName = existingDomainNames[i].name
        if (existingDomainName.endsWith('.')) existingDomainName = existingDomainName.substring(0, existingDomainName.length - 1)
        if (existingDomainName === tld) return existingDomainNames[i]
    }
    return undefined
}





