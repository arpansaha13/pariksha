import { shikiToMonaco } from '@shikijs/monaco'
import { defineStore, skipHydrate } from 'pinia'
import { createHighlighter } from 'shiki'
import type { editor } from 'monaco-editor'

export const useEditorStore = defineStore(editorStoreId, {
  // Create the highlighter, it can be reused
  state: () => ({
    ...skipHydrate({
      highlighter: null as Awaited<ReturnType<typeof createHighlighter>> | null,
    }),

    isEditorPrepared: false,
  }),
  getters: {
    getEditorOptions: (): editor.IStandaloneEditorConstructionOptions => ({
      minimap: { enabled: false },
      fontFamily: `'Fira Code', 'Cascadia Code', 'Monaco'`,
    }),
  },
  actions: {
    async createEditorHighlighter() {
      if (import.meta.server) {
        console.warn('createHighlighter is not meant to be called on server')
        return
      }

      if (this.highlighter !== null) {
        return Promise.resolve(this.highlighter)
      }

      this.highlighter = await createHighlighter({
        themes: ['light-plus'],
        langs: ['javascript'],
      })

      return this.highlighter
    },
    async prepareEditor() {
      if (import.meta.server) {
        console.warn('prepareMonacoEditor is not meant to be called on server')
        return
      }

      if (this.isEditorPrepared) return

      const monaco = useMonaco()!
      await this.createEditorHighlighter()

      // Register the languageIds first. Only registered languages will be highlighted.
      monaco.languages.register({ id: 'javascript' })
      shikiToMonaco(this.highlighter!, monaco)

      this.isEditorPrepared = true
    },
  },
})
