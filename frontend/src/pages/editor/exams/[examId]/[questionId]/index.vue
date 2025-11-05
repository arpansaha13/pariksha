<template>
  <NuxtLayout name="editor">
    <template #header="{ panelIsCollapsed, togglePanel }">
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
            @click="runCode(panelIsCollapsed, togglePanel)"
          />

          <UButton
            label="Save"
            :icon="isSaved ? 'lucide:cloud-check' : 'heroicons:cloud-arrow-up'"
            variant="soft"
            loading-auto
            @click="saveAnswer"
          />
        </div>
      </div>
    </template>

    <template #primary-sidebar>
      <DisplayCodingQuestion
        v-if="
          !isNullOrUndefined(questionData) &&
          questionData.type === QuestionType.CODING
        "
        :content="questionData.question"
        :test-cases="questionData.test_cases ?? []"
      />
    </template>

    <ClientOnly>
      <MonacoEditor
        v-if="editorStore.isEditorPrepared"
        v-model="examStore.codingAnswerStates[questionId].code"
        :lang="editorLang"
        :options="editorStore.getEditorOptions"
        class="h-full"
        @load="isEditorLoaded = true"
      />
    </ClientOnly>

    <template #panel="{ isCollapsed, toggle }">
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
                isCollapsed ? 'heroicons:chevron-up' : 'heroicons:chevron-down'
              "
              @click="toggle"
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
                  (questionData as QuestionCoding).question.input_definitions
                "
              />
            </template>
          </UTabs>
        </template>

        <template #results>
          <div
            v-if="isNullOrUndefined(engineRunResult)"
            class="grid size-full place-items-center"
          >
            <UEmpty
              variant="naked"
              icon="heroicons:code-bracket-16-solid"
              title="No execution results available"
              description="Run your code to see the output here"
            />
          </div>
          <!-- prettier-ignore -->
          <EditorTestCaseResults
            v-else
            :engine-run-result="engineRunResult"
            :question-data="(questionData as QuestionCoding)"
          />
        </template>
      </UTabs>
    </template>
  </NuxtLayout>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { TabsItem } from '@nuxt/ui'

definePageMeta({
  layout: false,
  middleware: 'store-previous-path',
})

const route = useRoute()
const examId = route.params.examId as ExamId
const questionId = route.params.questionId as QuestionId

const examStore = useExamStore()

// If the examId doesn't match, reset the store for the new exam
if (examStore.examId && examStore.examId !== examId) {
  examStore.$reset()
}

examStore.examId = examId

const {
  data: questionData,
  error: questionError,
  status: questionStatus,
} = await useExamQuestion(questionId)

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

const { data: boilerplateData } = await useQuestionCodingBoilerplate(
  questionId,
  1 as LanguageId
)

if (!examStore.codingAnswerStates[questionId]?.code) {
  const boilerplateAnswer: CodingAnswer = {
    code: boilerplateData.value?.code,
  }
  examStore.setAnswer(questionId, QuestionType.CODING, boilerplateAnswer)
}

const editorStore = useEditorStore()

onMounted(async () => {
  await editorStore.prepareEditor()
})

// _________________________BACK BUTTON___________________________
const previousPath = useState(UseStateKeys.PreviousPath)
const backButtonPath = computed(() => {
  if (previousPath.value) return previousPath.value
  if (isNullOrUndefined(questionData.value)) return undefined
  return `/exams/${examId}/attempt`
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
    slot: 'results',
  },
]

// __________________________TEST CASES___________________________
const testCaseTabItems = reactive<TestCase[]>(
  (questionData.value as QuestionCoding).test_cases!.map(testCase => ({
    inputs: testCase.inputs,
    expectedOutput: testCase.output,
  }))
)

// _________________________EDITOR STATE__________________________
const isEditorLoaded = ref(false)
const editorLang = ref(EditorLang.JAVASCRIPT)

const engineRunResult = ref<EngineRunResponse | null>(null)
async function runCode(panelIsCollapsed: boolean, togglePanel: () => void) {
  if (!isEditorLoaded.value) return
  engineRunResult.value = await engineRun({
    question_id: questionId,
    code: examStore.codingAnswerStates[questionId].code,
    environment: EngineEnv.NODE,
    test_cases: testCaseTabItems,
  })

  if (panelIsCollapsed) {
    togglePanel()
  }

  panelTabActive.value = PanelTabItemValue.RUN_RESULTS
}

// _______________________MANUAL SAVE ANSWER______________________
const isSaved = ref(false)
async function saveAnswer() {
  await Promise.all(examStore.saveUpdatedAnswers())
  isSaved.value = true
  useTimeoutFn(() => {
    isSaved.value = false
  }, 3000)
}

// _______________________AUTO SAVE ANSWER________________________
const { $api } = useNuxtApp()

watchImmediate(
  () => route.params.questionId as QuestionId,
  async qid => {
    if (import.meta.server) return
    if (isNullOrUndefined(qid)) return

    if (examStore.answerFetched[qid]) return
    examStore.answerFetched[qid] = true

    const data = await $api<AnswerMinimal>(
      `/api/exams/${examId}/questions/${qid}/answer`
    )

    if (isNullOrUndefined(data.answer)) return
    if (data.question_id !== qid) {
      examStore.answerFetched[qid] = false
      return
    }

    examStore.setAnswer(qid, QuestionType.CODING, data.answer)
  }
)

useIntervalFn(
  examStore.saveUpdatedAnswers,
  AUTO_SAVE_EXAM_ANSWER_INTERVAL_SECONDS * 1000
)
</script>
