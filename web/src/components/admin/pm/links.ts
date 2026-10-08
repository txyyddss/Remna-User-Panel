export function pmTopicLink(chatId: string, topicId: string | null, messageId?: string | null): string | null {
  if (!/^-100\d+$/.test(chatId) || !topicId || !/^\d+$/.test(topicId)) return null
  const group = chatId.slice(4)
  return messageId && /^\d+$/.test(messageId)
    ? `https://t.me/c/${group}/${topicId}/${messageId}`
    : `https://t.me/c/${group}/${topicId}`
}
