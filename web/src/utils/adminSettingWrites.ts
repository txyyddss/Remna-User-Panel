// Dependent integration toggles must follow their validated keys/destination.
export async function writeAdminSettings(values: [string, string][], write: (key: string, value: string) => Promise<unknown>): Promise<void> {
  const toggles = new Set(['captcha.turnstile.enabled', 'telegram.pm.enabled'])
  const disabled = values.filter(([key, value]) => toggles.has(key) && value === 'false')
  const configured = values.filter(([key]) => !toggles.has(key))
  const enabled = values.filter(([key, value]) => toggles.has(key) && value === 'true')
  await Promise.all(disabled.map(([key, value]) => write(key, value)))
  await Promise.all(configured.map(([key, value]) => write(key, value)))
  for (const [key, value] of enabled) await write(key, value)
}
