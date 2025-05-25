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
          loading-auto
          @click="runCode"
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

        <SplitterPanel id="splitter-group-1-panel-2" :min-size="20">
          <SplitterGroup id="splitter-group-2" direction="vertical">
            <SplitterPanel
              id="splitter-group-2-panel-1"
              :min-size="20"
              class="splitter-panel"
            >
              <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
                <MonacoEditor
                  v-if="editorStore.isEditorPrepared"
                  v-model="editorCode"
                  :lang="editorLang"
                  :options="editorStore.getEditorOptions"
                  class="h-full"
                  @load="isEditorLoaded = true"
                />
              </UCard>
            </SplitterPanel>

            <SplitterResizeHandle
              id="splitter-group-2-resize-handle-1"
              class="group flex h-2 flex-col outline-none"
            >
              <div
                class="group-hover:bg-primary-500 group-focus:bg-primary-500 my-auto h-0.5 w-full transition-colors delay-150"
              />
            </SplitterResizeHandle>

            <SplitterPanel id="splitter-group-2-panel-2" class="splitter-panel">
              <UCard
                :ui="{
                  root: 'h-full overflow-auto',
                  body: 'h-full overflow-auto',
                }"
              >
                <DisplayCodeBlock
                  v-if="engineRunResult?.stdout"
                  preserve-white-space
                >
                  <template #header> Stdout </template>
                  <p>{{ engineRunResult.stdout }}</p>
                </DisplayCodeBlock>

                <DisplayCodeBlock
                  v-if="engineRunResult?.stderr"
                  color="error"
                  preserve-white-space
                >
                  <template #header> Stderr </template>
                  <p class="text-error-400">
                    {{ engineRunResult.stderr }}
                  </p>
                </DisplayCodeBlock>
              </UCard>
            </SplitterPanel>
          </SplitterGroup>
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

const isEditorLoaded = ref(false)
const editorLang = ref(EditorLang.JAVASCRIPT)
const editorCode = ref(`function helloDocker() {
  const message = "hello docker!!"
  console.log(message)
}
helloDocker()
`)

const engineRunResult = ref<EngineRunResult | null>(null)
async function runCode() {
  if (!isEditorLoaded.value) return
  engineRunResult.value = await engineRun({
    code: editorCode.value,
    environment: EngineEnv.NODE,
  })
}
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
