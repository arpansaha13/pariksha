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
      :id="Splitter.GROUP_1_ID"
      direction="horizontal"
      @layout="splitterGroup1Layout = $event"
    >
      <SplitterPanel
        :id="Splitter.PRIMARY_SIDE_BAR"
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
        :id="Splitter.GROUP_1_RESIZE_HANDLE_1_ID"
        class="group w-2 outline-none"
      >
        <div
          class="group-hover:bg-primary-500 group-focus:bg-primary-500 mx-auto h-full w-0.5 transition-colors delay-150"
        />
      </SplitterResizeHandle>

      <SplitterPanel
        :id="Splitter.GROUP_1_PANEL_2_ID"
        :default-size="splitterGroup1Layout?.[1]"
        :min-size="20"
      >
        <SplitterGroup
          :id="Splitter.GROUP_2_ID"
          ref="splitterGroup2Ref"
          direction="vertical"
          @layout="splitterGroup2Layout = $event"
        >
          <SplitterPanel
            :id="Splitter.GROUP_2_PANEL_1_ID"
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
            :id="Splitter.GROUP_2_RESIZE_HANDLE_1_ID"
            class="group flex h-2 flex-col outline-none"
          >
            <div
              class="group-hover:bg-primary-500 group-focus:bg-primary-500 my-auto h-0.5 w-full transition-colors delay-150"
            />
          </SplitterResizeHandle>

          <SplitterPanel
            :id="Splitter.GROUP_2_PANEL_2_ID"
            ref="panelRef"
            collapsible
            :default-size="splitterGroup2Layout?.[1]"
            :min-size="splitterGroup2Panel2CollapsedSize"
            :collapsed-size="splitterGroup2Panel2CollapsedSize"
            class="splitter-panel"
          >
            <UCard
              :ui="{
                root: 'h-full flex flex-col',
                body: 'h-full p-0!',
              }"
            >
              <UTabs
                v-model="panelTabActive"
                :items="panelTabItems"
                size="sm"
                variant="link"
                class="w-full"
                :ui="{
                  root: 'h-full',
                  content: 'grow px-4 pb-4 overflow-auto',
                }"
              >
                <template #list-trailing>
                  <div class="ml-auto">
                    <UButton
                      size="sm"
                      color="neutral"
                      variant="ghost"
                      :icon="
                        panelRef?.isCollapsed
                          ? 'heroicons:chevron-up'
                          : 'heroicons:chevron-down'
                      "
                      @click="toggleSplitterPanel(panelRef)"
                    />
                  </div>
                </template>

                <template #test-cases>
                  <UTabs :items="testCaseTabItems" size="sm" variant="link">
                    <template #default="{ index: testCaseIdx }">
                      Test case {{ testCaseIdx + 1 }}
                    </template>

                    <template #content="{ index: testCaseIdx }">
                      <EditorTestCaseForm
                        v-model:test-case="testCaseTabItems[testCaseIdx]"
                        :test-case-idx="testCaseIdx"
                        :input-definitions="
                          (questionData as QuestionCoding).question
                            .input_definitions
                        "
                      />
                    </template>
                  </UTabs>
                </template>

                <template #run-results>
                  <div
                    v-if="isNullOrUndefined(engineRunResult)"
                    class="grid size-full place-items-center"
                  >
                    <EmptyState
                      icon="heroicons:code-bracket-16-solid"
                      title="No execution results available"
                      description="Run your code to see the output here"
                    />
                  </div>

                  <EditorRunResult v-else :run-result="engineRunResult" />
                </template>
              </UTabs>
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
import type { TabsItem } from '@nuxt/ui'

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

const editorStore = useEditorStore()

onMounted(async () => {
  await editorStore.prepareEditor()
})

enum Splitter {
  GROUP_1_ID = 'splitter-group-1',
  PRIMARY_SIDE_BAR = 'primary-side-bar', // splitter-group-1-panel-1
  GROUP_1_PANEL_2_ID = 'splitter-group-1-panel-2',
  GROUP_1_RESIZE_HANDLE_1_ID = 'splitter-group-1-resize-handle-1',

  GROUP_2_ID = 'splitter-group-2',
  GROUP_2_PANEL_1_ID = 'splitter-group-2-panel-1',
  GROUP_2_PANEL_2_ID = 'splitter-group-2-panel-2',
  GROUP_2_RESIZE_HANDLE_1_ID = 'splitter-group-2-resize-handle-1',
}

// ______________________SPLITTER LAYOUT SSR______________________
const splitterGroup1Layout = useCookie<number[]>(
  'editor:splitter-group-1-layout'
)
const splitterGroup2Layout = useCookie<number[]>(
  'editor:splitter-group-2-layout'
)

// ____________CALCULATE GROUP-2 PANEL-2 COLLAPSED SIZE___________
const splitterGroup2Ref = ref<InstanceType<typeof SplitterGroup>>()
const panelRef = ref<InstanceType<typeof SplitterPanel>>()

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

// ________________SPLITTER PANEL EXPAND/COLLAPSE_________________
function useSplitterToggleCollapse() {
  const isPanelInitiallyCollapsed = ref({} as Record<Splitter, boolean>)

  const { stop } = watchEffect(async () => {
    if (panelRef.value?.isCollapsed) {
      isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID] = true
      stop()
    }
  })

  function toggleSplitterPanel(
    panelInstance: InstanceType<typeof SplitterPanel> | undefined
  ) {
    if (isNullOrUndefined(panelInstance)) return

    if (panelInstance.isExpanded) {
      panelInstance.collapse()
      return
    }

    if (isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID]) {
      panelInstance.resize(45)
      isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID] = false
    } else {
      panelInstance.expand()
    }
  }
  return toggleSplitterPanel
}

const toggleSplitterPanel = useSplitterToggleCollapse()

// _________________________BACK BUTTON___________________________
const previousPath = useState(UseStateKeys.PreviousPath)
const backButtonPath = computed(() => {
  if (previousPath.value) return previousPath.value
  if (isNullOrUndefined(questionData.value)) return undefined
  return `/papers/${questionData.value.paper_id}`
})

// ______________________PANEL TABS_____________________
enum PanelTabItemValue {
  TEST_CASES = 1,
  RUN_RESULTS = 2,
}

const panelTabActive = ref(PanelTabItemValue.TEST_CASES)

const panelTabItems: TabsItem[] = [
  {
    value: PanelTabItemValue.TEST_CASES,
    label: 'Test cases',
    slot: 'test-cases',
  },
  {
    value: PanelTabItemValue.RUN_RESULTS,
    label: 'Run results',
    slot: 'run-results',
  },
]

// __________________________TEST CASES___________________________
const testCaseTabItems = reactive<TestCase[]>(
  (questionData.value!.question as QuestionCodingContent).test_cases!.map(
    testCase => ({
      inputs: testCase.inputs,
      expectedOutput: testCase.output,
    })
  )
)

// _________________________EDITOR STATE__________________________
const isEditorLoaded = ref(false)
const editorLang = ref(EditorLang.JAVASCRIPT)
const editorCode = ref(`function solve(a, b) {
  return parseInt(a) + parseInt(b)
}`)

const engineRunResult = ref<EngineRunResult | null>(null)
async function runCode() {
  if (!isEditorLoaded.value) return
  engineRunResult.value = await engineRun({
    code: editorCode.value,
    environment: EngineEnv.NODE,
    testCases: testCaseTabItems,
  })

  if (panelRef.value?.isCollapsed) {
    toggleSplitterPanel(panelRef.value)
  }

  panelTabActive.value = PanelTabItemValue.RUN_RESULTS
}
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
