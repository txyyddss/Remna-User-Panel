import { z } from 'zod'

const rewardSchema = z.discriminatedUnion('kind', [
  z.object({ kind: z.literal('none') }),
  z.object({ kind: z.literal('coupon'), couponId: z.string().min(1) }),
  z.object({ kind: z.literal('txb'), txbMinor: z.number().int().positive().max(100_000_000_000) }),
  z.object({ kind: z.literal('subscription_extension'), extensionDays: z.number().int().min(1).max(3650) }),
])

const tierSchema = z.object({
  id: z.string(), name: z.string().trim().min(1).max(48), threshold: z.number().int().nonnegative(), enabled: z.boolean(),
  commissionEnabled: z.boolean(), commissionBps: z.number().int().min(0).max(10_000), reward: rewardSchema,
}).refine((tier) => tier.commissionEnabled || tier.commissionBps === 0, { path: ['commissionBps'] })

export const affiliateFormSchema = z.object({ tiers: z.array(tierSchema).min(1).max(50) }).superRefine(({ tiers }, context) => {
  const enabled = tiers.filter((tier) => tier.enabled)
  if (!enabled.length || enabled[0]?.threshold !== 0) context.addIssue({ code: 'custom', path: ['tiers'], message: 'threshold' })
  for (let index = 1; index < enabled.length; index += 1) {
    if (enabled[index]!.threshold <= enabled[index - 1]!.threshold) context.addIssue({ code: 'custom', path: ['tiers', index, 'threshold'], message: 'ordering' })
  }
  if (enabled[0]?.reward.kind !== 'none') context.addIssue({ code: 'custom', path: ['tiers', 0, 'reward'], message: 'starterReward' })
})
