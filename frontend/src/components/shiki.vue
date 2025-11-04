<template>
  <ClientOnly>
    <div class="overflow-auto" v-html="highlightedCode"/>
  </ClientOnly>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

const props = defineProps<{
  code: string
}>()

const editorStore = useEditorStore()
await editorStore.createEditorHighlighter()

const highlightedCode = computed(() => {
  if (isNullOrUndefined(editorStore.highlighter)) return ''
  return editorStore.highlighter!.codeToHtml(props.code, {
    lang: 'javascript',
    theme: 'light-plus',
  })
})
</script>
