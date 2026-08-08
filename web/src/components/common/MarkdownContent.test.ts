import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import MarkdownContent from './MarkdownContent.vue'

describe('MarkdownContent', () => {
  it('renders safe Markdown while removing raw HTML and unsafe protocols', () => {
    const wrapper = mount(MarkdownContent, {
      props: {
        source: '**Safe** <img src=x onerror=alert(1)> [docs](https://example.com) [bad](javascript:alert(1))',
      },
    })

    expect(wrapper.get('strong').text()).toBe('Safe')
    expect(wrapper.find('img').exists()).toBe(false)
    const safeLink = wrapper.get('a[href="https://example.com"]')
    expect(safeLink.attributes('target')).toBe('_blank')
    expect(safeLink.attributes('rel')).toContain('noopener')
    expect(wrapper.findAll('a')).toHaveLength(1)
    expect(wrapper.find('[onerror]').exists()).toBe(false)
  })
})
