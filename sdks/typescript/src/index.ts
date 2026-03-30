export { APIClient } from './client.js'
export { buildCSS, contrastRatio, validateContrasts } from './tokens.js'
export { renderPreview, renderCustomFieldsHTML } from './preview.js'
export { loadConfig, loadLocalTemplates, loadLocalStrings, computeDiff, push, pull } from './sync.js'
export type {
  PageType, PageTemplate, LanguageString, AuthTokens, BrandingConfig,
  SDKConfig, DesignTokens, ProjectConfig, CustomFieldConfig, SyncDiff, ContrastResult,
} from './types.js'

import { APIClient } from './client.js'
import * as tokens from './tokens.js'
import * as preview from './preview.js'
import * as sync from './sync.js'
import type { SDKConfig, DesignTokens, ProjectConfig } from './types.js'

export class CPIAuth {
  readonly client: APIClient
  readonly templates: APIClient
  readonly strings: {
    list: APIClient['listStrings']
    set: APIClient['upsertString']
    delete: APIClient['deleteString']
  }
  readonly tokens: { buildCSS: typeof tokens.buildCSS; validate: typeof tokens.validateContrasts }
  readonly preview: { render: typeof preview.renderPreview }
  readonly sync: {
    diff: (config: ProjectConfig) => ReturnType<typeof sync.computeDiff>
    push: (config: ProjectConfig, opts?: Parameters<typeof sync.push>[2]) => ReturnType<typeof sync.push>
    pull: (config: ProjectConfig) => ReturnType<typeof sync.pull>
  }

  constructor(config: SDKConfig) {
    this.client = new APIClient(config)

    // Convenience aliases
    this.templates = this.client
    this.strings = {
      list: this.client.listStrings.bind(this.client),
      set: this.client.upsertString.bind(this.client),
      delete: this.client.deleteString.bind(this.client),
    }
    this.tokens = {
      buildCSS: tokens.buildCSS,
      validate: tokens.validateContrasts,
    }
    this.preview = {
      render: preview.renderPreview,
    }
    this.sync = {
      diff: (cfg) => sync.computeDiff(this.client, cfg),
      push: (cfg, opts) => sync.push(this.client, cfg, opts),
      pull: (cfg) => sync.pull(this.client, cfg),
    }
  }

  async login(email: string, password: string) {
    return this.client.login(email, password)
  }
}
