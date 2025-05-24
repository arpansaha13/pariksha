<template>
  <main class="flex h-screen w-screen flex-col gap-2 px-4 py-2">
    <div class="grid grid-cols-3 gap-4">
      <div>
        <UButton
          label="Back"
          icon="i-heroicons-arrow-uturn-left"
          color="neutral"
          variant="soft"
          :to="backButtonPath"
        />
      </div>

      <div class="flex justify-center gap-2">
        <UButton
          label="Run"
          icon="heroicons:play"
          color="neutral"
          variant="soft"
        />

        <UButton label="Save" icon="heroicons:cloud-arrow-up" variant="soft" />
      </div>
    </div>

    <ClientOnly>
      <SplitterGroup
        id="splitter-group-1"
        direction="horizontal"
        auto-save-id="editor-splitter-group-1-save"
      >
        <SplitterPanel
          id="splitter-group-1-panel-1"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard
            :ui="{
              root: 'overflow-hidden h-full',
              body: 'overflow-y-auto h-full',
            }"
          >
            <DisplayCodingQuestion
              v-if="
                !isNullOrUndefined(questionData) &&
                questionData.type === QuestionType.CODING
              "
              :content="questionData.question"
            />
          </UCard>
        </SplitterPanel>

        <SplitterResizeHandle
          id="splitter-group-1-resize-handle-1"
          class="group w-2 outline-none"
        >
          <div
            class="group-hover:bg-primary-500 group-focus:bg-primary-500 mx-auto h-full w-0.5 transition-colors delay-150"
          />
        </SplitterResizeHandle>

        <SplitterPanel
          id="splitter-group-1-panel-2"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
            <MonacoEditor
              v-if="editorStore.isEditorPrepared"
              v-model="value"
              :lang="editorLang"
              :options="editorStore.getEditorOptions"
              class="h-full"
            />
          </UCard>
        </SplitterPanel>
      </SplitterGroup>
    </ClientOnly>
  </main>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

definePageMeta({
  layout: 'blank',
  middleware: 'store-previous-path',
})

const route = useRoute()
const questionId = parseInt(route.params.questionId as string) as QuestionId

const {
  data: questionData,
  error: questionError,
  status: questionStatus,
} = await usePaperQuestion(questionId)

const previousPath = useState(UseStateKeys.PreviousPath)

const backButtonPath = computed(() => {
  if (previousPath.value) return previousPath.value
  if (isNullOrUndefined(questionData.value)) return undefined
  return `/papers/${questionData.value.paper_id}`
})

const editorStore = useEditorStore()

await callOnce(
  () => {
    if (questionStatus.value === 'error' && questionError.value) {
      throw createError({
        statusCode: questionError.value.statusCode,
        message: 'You do not have access to this paper.',
      })
    }
    if (
      isNullOrUndefined(questionData.value) ||
      questionData.value.type !== QuestionType.CODING
    ) {
      throw createError({
        statusCode: HttpStatus.NOT_FOUND,
        message: 'We could not find the question you are looking for.',
      })
    }
  },
  { mode: 'navigation' }
)

onMounted(async () => {
  await editorStore.prepareEditor()
})

const editorLang = ref(EditorLang.JAVASCRIPT)
const value = ref(`function getColor() {
  const a = ""
}`)
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
