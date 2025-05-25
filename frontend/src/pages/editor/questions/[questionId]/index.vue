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

    <SplitterGroup
      id="splitter-group-1"
      direction="horizontal"
      @layout="splitterGroup1Layout = $event"
    >
      <SplitterPanel
        id="splitter-group-1-panel-1"
        :default-size="splitterGroup1Layout?.[0]"
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
        :default-size="splitterGroup1Layout?.[1]"
        :min-size="20"
      >
        <SplitterGroup
          id="splitter-group-2"
          ref="splitterGroup2Ref"
          direction="vertical"
          @layout="splitterGroup2Layout = $event"
        >
          <SplitterPanel
            id="splitter-group-2-panel-1"
            :default-size="splitterGroup2Layout?.[0]"
            :min-size="20"
            class="splitter-panel"
          >
            <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
              <ClientOnly>
                <MonacoEditor
                  v-if="editorStore.isEditorPrepared"
                  v-model="editorCode"
                  :lang="editorLang"
                  :options="editorStore.getEditorOptions"
                  class="h-full"
                  @load="isEditorLoaded = true"
                />
              </ClientOnly>
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

          <SplitterPanel
            id="splitter-group-2-panel-2"
            ref="splitterGroup2Panel2Ref"
            collapsible
            :default-size="splitterGroup2Layout?.[1]"
            :min-size="splitterGroup2Panel2CollapsedSize"
            :collapsed-size="splitterGroup2Panel2CollapsedSize"
            class="splitter-panel"
          >
            <UCard
              :ui="{
                root: 'h-full flex flex-col',
                body: 'grow overflow-auto',
                header: 'p-1! flex items-center justify-between',
              }"
            >
              <template #header>
                <div class="px-1.5">
                  <p class="text-sm">Run Results</p>
                </div>

                <div>
                  <UButton
                    size="sm"
                    color="neutral"
                    variant="ghost"
                    :icon="
                      splitterGroup2Panel2Ref?.isCollapsed
                        ? 'heroicons:chevron-up'
                        : 'heroicons:chevron-down'
                    "
                    @click="
                      splitterGroup2Panel2Ref?.isCollapsed
                        ? splitterGroup2Panel2Ref?.expand()
                        : splitterGroup2Panel2Ref?.collapse()
                    "
                  />
                </div>
              </template>

              <div
                v-if="isNullOrUndefined(engineRunResult)"
                class="grid h-full place-items-center"
              >
                <EmptyState
                  icon="heroicons:code-bracket-16-solid"
                  title="No execution results available"
                  description="Run your code to see the output here"
                />
              </div>

              <DisplayCodeBlock
                v-else-if="engineRunResult.stdout"
                preserve-white-space
              >
                <template #header> Stdout </template>
                <p>{{ engineRunResult.stdout }}</p>
              </DisplayCodeBlock>

              <DisplayCodeBlock
                v-else-if="engineRunResult.stderr"
                color="error"
                preserve-white-space
              >
                <template #header> Stderr </template>
                <p class="text-error-500">
                  {{ engineRunResult.stderr }}
                </p>
              </DisplayCodeBlock>
            </UCard>
          </SplitterPanel>
        </SplitterGroup>
      </SplitterPanel>
    </SplitterGroup>
  </main>
</template>

<script lang="ts" setup>
import type { SplitterGroup, SplitterPanel } from '#components'
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

const previousPath = useState(UseStateKeys.PreviousPath)
const editorStore = useEditorStore()

const splitterGroup1Layout = useCookie<number[]>(
  'editor:splitter-group-1-layout'
)
const splitterGroup2Layout = useCookie<number[]>(
  'editor:splitter-group-2-layout'
)

const splitterGroup2Ref = ref<InstanceType<typeof SplitterGroup>>()
const splitterGroup2Panel2Ref = ref<InstanceType<typeof SplitterPanel>>()

const splitterGroup2Panel2CollapsedSize = computed(() => {
  if (!splitterGroup2Ref.value) return 0

  const el = splitterGroup2Ref.value.$el as HTMLDivElement
  const splitterGroup2Height = el.clientHeight
  const splitterGroup2Panel2HeaderHeight = 36.8 // Update this if header height changes
  const splitterGroup2ResizeHandlerHeight = 0 // Update this if resize-handler height changes
  const effectiveAvailableHeight =
    splitterGroup2Panel2HeaderHeight - splitterGroup2ResizeHandlerHeight
  const collapsedSizePercent =
    (effectiveAvailableHeight / splitterGroup2Height) * 100

  return Math.floor(collapsedSizePercent * 100) / 100 // 2 decimal places
})

const backButtonPath = computed(() => {
  if (previousPath.value) return previousPath.value
  if (isNullOrUndefined(questionData.value)) return undefined
  return `/papers/${questionData.value.paper_id}`
})

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

  if (splitterGroup2Panel2Ref.value?.isCollapsed) {
    splitterGroup2Panel2Ref.value.expand()
  }
}
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
