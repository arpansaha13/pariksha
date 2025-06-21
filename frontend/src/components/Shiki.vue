<template>
  <ClientOnly>
    <div v-html="highlightedCode" class="overflow-auto"></div>
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
