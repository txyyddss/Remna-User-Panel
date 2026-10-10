import { ipLookupApi, type IPLookupLiveDetails } from '@/api/ipLookup'
import { ApiError } from '@/api/http'

export function installLiveIPAudit(params: URLSearchParams) {
  let calls = 0
  const requested: string[] = []
  ipLookupApi.details = async (ip, signal) => {
    calls++; requested.push(ip)
    if (params.get('live') === 'slow') await new Promise<void>((resolve, reject) => {
      const timer = setTimeout(resolve, 1800)
      signal?.addEventListener('abort', () => { clearTimeout(timer); reject(new DOMException('Aborted', 'AbortError')) }, { once: true })
    })
    if (params.get('live') === 'error') throw new ApiError(429, { code: 'IP_DETAILS_RATE_LIMITED', message: 'Constructed rate limit' })
    const now = new Date().toISOString()
    const meta = { status: 'success' as const, source: 'ripe:looking-glass', observedAt: now, errorCode: '' }
    const asns = [174, 3257, 2914, 1299, 6461, 6939, 3356, 7018, 9002, 12956, 6453, 3491, 3549, 5511]
    const names: Record<number, string> = { 13335: 'Cloudflare, Inc.', 174: 'Cogent Communications', 3257: 'GTT Communications', 2914: 'NTT DATA', 1299: 'Arelion', 6461: 'Zayo', 6939: 'Hurricane Electric', 3356: 'Lumen', 7018: 'AT&T', 9002: 'RETN', 12956: 'Telefónica', 6453: 'Tata Communications', 3491: 'PCCW Global', 3549: 'Level 3', 5511: 'Orange' }
    const result: IPLookupLiveDetails = {
      ip, fetchedAt: now,
      topology: { ...meta, prefix: ip.includes(':') ? '2606:4700:4700::/48' : '8.8.8.0/24', origins: [13335], nodes: [13335, ...asns, 48266].map(asn => ({ asn, name: names[asn] ?? 'Catixs' })), edges: [...asns.map((asn, index) => ({ from: asn, to: 13335, relationship: index < 4 ? 'upstream' as const : 'observed' as const })), { from: 48266, to: 174, relationship: 'observed' }], paths: asns.map((asn, index) => ({ asPath: index === 0 ? [48266, asn, 13335, 13335] : [asn, 13335], collector: `RRC${String(index).padStart(2, '0')}`, exchange: index === 0 ? 'LINX / LONAP' : '', observedAt: now })), truncated: false, relationshipMeta: { ...meta, source: 'caida', observedAt: '2026-10-01' } },
      asns: [{ asn: 13335, name: names[13335]!, registeredAt: '2010-07-14T00:00:00Z', rir: 'ARIN', announced: true, ipv4Prefixes: 2676, ipv6Prefixes: 3049, observedNeighbors: 1543, peers: 362, upstreams: 109, sources: { registration: { ...meta, source: 'ripe:rir' }, rir: { ...meta, source: 'ripe:rir' }, routing: { ...meta, source: 'ripe:routing-status' }, relationships: { ...meta, source: 'caida', observedAt: '2026-10-01' } }, traffic: {
        botHuman: { ...meta, source: 'cloudflare', values: { bot: 32.7, human: 67.3 }, startTime: '2026-10-03', endTime: '2026-10-10' },
        devices: { ...meta, source: 'cloudflare', values: { desktop: 42.3, mobile: 55.8, other: 1.9 }, startTime: '2026-10-03', endTime: '2026-10-10' },
        ipVersion: { ...meta, source: 'cloudflare', values: { IPv4: 58.4, IPv6: 41.6 }, startTime: '2026-10-03', endTime: '2026-10-10' },
      } }],
      scores: { ...meta, source: 'ipapi', company: { ratio: 0.0003, label: 'Low' }, asn: { ratio: 0, label: 'Very Low' } },
      block: { ...meta, source: 'abuseipdb', prefix: '8.8.8.0/24', reportedAddresses: 0, addressCapacity: '256', windowDays: 30 },
      ptr: { ...meta, source: 'dns', domains: ['dns.google', 'resolver.google'] },
      shodan: { ...meta, source: 'shodan:internetdb', observedAt: '', hostnames: ['dns.google'], ports: [53, 443, 853] },
    }
    if (params.get('live') === 'empty') {
      result.topology = { ...result.topology, status: 'empty', prefix: '', origins: [], nodes: [], edges: [], paths: [] }
      result.asns = []; result.ptr = { ...result.ptr, status: 'empty', domains: [] }
      result.shodan = { ...result.shodan, status: 'empty', hostnames: [], ports: [] }
      result.scores = { ...result.scores, status: 'unconfigured', company: { ratio: null, label: '' }, asn: { ratio: null, label: '' } }
      result.block = { ...result.block, status: 'empty', reportedAddresses: null, addressCapacity: '' }
    }
    if (params.get('live') === 'partial') {
      result.block = { ...result.block, status: 'restricted', reportedAddresses: null, errorCode: 'IP_DETAILS_PLAN_RESTRICTED' }
      result.ptr = { ...result.ptr, status: 'error', domains: [], errorCode: 'IP_DETAILS_UNAVAILABLE' }
      result.shodan = { ...result.shodan, status: 'error', hostnames: [], ports: [], errorCode: 'IP_DETAILS_UNAVAILABLE' }
      for (const summary of Object.values(result.asns[0]!.traffic)) { summary.status = 'unconfigured'; summary.values = {} }
    }
    if (params.get('live') === 'moas') {
      result.topology.origins.push(48266)
      result.topology.paths.push({ asPath: [3257, 48266], collector: 'RRC21', exchange: '', observedAt: now })
      result.topology.edges.push({ from: 3257, to: 48266, relationship: 'upstream' })
      result.asns.push({ ...result.asns[0]!, asn: 48266, name: 'Catixs', rir: 'RIPE NCC' })
    }
    return result
  }
  return () => ({ calls, requested })
}
