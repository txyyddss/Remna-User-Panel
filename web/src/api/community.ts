import type { CommunityMembership, CommunitySpace, JoinInvite } from './types'
import { request } from './http'

// communityApi owns membership facts and one-space Telegram invite requests.
export const communityApi = {
  checkCommunityMembership: () => request<CommunityMembership>('/api/v1/community/membership/check', { method: 'POST' }),
  createCommunityInvite: (space: CommunitySpace) => request<JoinInvite>(`/api/v1/community/invites/${space}`, { method: 'POST' }),
}
