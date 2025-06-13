import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import App from '@/App.vue'

describe('App.vue', () => {
  it('renders properly', () => {
    const pinia = createPinia()
    const wrapper = mount(App, {
      global: {
        plugins: [pinia]
      }
    })
    expect(wrapper.text()).toContain('Engraving Processor Pro')
  })

  it('displays version', () => {
    const pinia = createPinia()
    const wrapper = mount(App, {
      global: {
        plugins: [pinia]
      }
    })
    expect(wrapper.text()).toContain('v1.0.0')
  })
})