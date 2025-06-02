<template>
  <UModal
    title="Define parameters"
    description="Define the number and type of inputs and output for the program"
    :ui="{ body: 'divide-y divide-default' }"
    @after:leave="emit('after:leave')"
  >
    <UButton
      icon="heroicons:variable"
      size="sm"
      color="neutral"
      variant="subtle"
      label="Define parameters"
    />

    <template #body>
      <UAlert
        v-if="hasTestCases"
        color="warning"
        variant="subtle"
        description="Updating input/output definitions will clear existing test cases."
        icon="heroicons:exclamation-circle"
      />

      <UForm
        v-for="(
          inputDefinition, inputIdx
        ) in codingQuestionContent.input_definitions"
        :key="inputIdx"
        :state="codingQuestionContent.input_definitions[inputIdx]"
        class="space-y-1.5 py-2 first:pt-0 last:pb-0"
      >
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-bold">Input {{ inputIdx + 1 }}</h3>

          <UButton
            icon="heroicons:trash"
            size="sm"
            color="error"
            variant="subtle"
            class="flex"
            :disabled="
              codingQuestionContent.input_definitions.length <=
              MIN_CODING_INPUTS_COUNT
            "
            @click="removeInputDefinition(inputIdx)"
          />
        </div>

        <UFormField
          label="Type"
          description="Data type of the input argument"
          :name="`input-definition-${inputIdx + 1}-type`"
          required
        >
          <UButtonGroup>
            <UBadge
              color="neutral"
              variant="subtle"
              size="lg"
              :label="inputDefinition.variable_name"
            />

            <USelect
              v-model="inputDefinition.type"
              :items="inputTypeSelectItems"
              required
              class="w-48"
              @change="
                onParameterTypeChange(
                  codingQuestionContent.input_definitions[inputIdx]
                )
              "
            />
          </UButtonGroup>
        </UFormField>

        <template
          v-if="
            inputDefinition.type ===
            QuestionCodingContentCompositeInputTypes.ARRAY
          "
        >
          <UFormField
            v-for="(subInputItem, subInputItemIdx) in inputDefinition.items"
            :key="subInputItemIdx"
            label="Array item type"
            description="Data type of each array item"
            :name="`input-definition-${inputIdx + 1}-sub-type`"
            required
          >
            <USelect
              v-model="subInputItem.type"
              :items="inputSubTypeSelectItems"
              required
              class="w-48"
            />
          </UFormField>
        </template>
      </UForm>

      <UForm
        :state="codingQuestionContent.output_definition"
        class="space-y-1.5 py-2 first:pt-0 last:pb-0"
      >
        <h3 class="text-sm font-bold">Output</h3>

        <UFormField
          label="Type"
          description="Data type of the output. This will be used as the return type."
          name="output-definition-type"
          required
        >
          <USelect
            v-model="codingQuestionContent.output_definition.type"
            :items="inputTypeSelectItems"
            required
            class="w-48"
            @change="
              onParameterTypeChange(codingQuestionContent.output_definition)
            "
          />
        </UFormField>

        <template
          v-if="
            codingQuestionContent.output_definition.type ===
            QuestionCodingContentCompositeInputTypes.ARRAY
          "
        >
          <UFormField
            v-for="(subInputItem, subInputItemIdx) in codingQuestionContent
              .output_definition.items"
            :key="subInputItemIdx"
            label="Array item type"
            description="Data type of each array item"
            :name="`output-definition-sub-type`"
            required
          >
            <USelect
              v-model="subInputItem.type"
              :items="inputSubTypeSelectItems"
              required
              class="w-48"
            />
          </UFormField>
        </template>
      </UForm>
    </template>

    <template #footer>
      <UButton
        color="neutral"
        variant="subtle"
        label="Add input"
        :disabled="
          codingQuestionContent.input_definitions.length >=
          MAX_CODING_INPUTS_COUNT
        "
        @click="addInputDefinition"
      />
    </template>
  </UModal>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

defineProps<{
  hasTestCases: boolean
}>()

const codingQuestionContent = defineModel<
  Pick<QuestionCodingContent, 'input_definitions' | 'output_definition'>
>('coding-question-content', {
  required: true,
})

const emit = defineEmits(['after:leave'])

// ________________________INPUT DEFINITIONS________________________
function addInputDefinition() {
  const inputDefinitions = codingQuestionContent.value.input_definitions

  // Prevent adding more than 5 inputs
  if (inputDefinitions.length < MAX_CODING_INPUTS_COUNT) {
    inputDefinitions.push({
      variable_name: getDefaultCodingQuestionInputVariableName(
        inputDefinitions.length + 1
      ),
      type: 0 as QuestionCodingContentInputTypes,
    })
  }
}

function removeInputDefinition(idx: number) {
  const inputDefinitions = codingQuestionContent.value.input_definitions

  // Prevent deleting options when there are only 2 left
  if (inputDefinitions.length > MIN_CODING_INPUTS_COUNT) {
    inputDefinitions.splice(idx, 1)
  }
}

// _____________________INPUT TYPES SELECT ITEMS____________________
const primitiveInputTypes = [
  {
    type: 'label',
    label: 'Primitive',
  },
  {
    label: 'Number',
    value: QuestionCodingContentPrimitiveInputTypes.NUMBER,
  },
  {
    label: 'String',
    value: QuestionCodingContentPrimitiveInputTypes.STRING,
  },
  {
    label: 'Boolean',
    value: QuestionCodingContentPrimitiveInputTypes.BOOLEAN,
  },
]

const compositeInputTypes = [
  {
    type: 'label',
    label: 'Composite',
  },
  {
    label: 'Array',
    value: QuestionCodingContentCompositeInputTypes.ARRAY,
  },
]

const inputTypeSelectItems = ref([primitiveInputTypes, compositeInputTypes])

/** Sub types for composite input types */
const inputSubTypeSelectItems = ref([primitiveInputTypes])

// ________________EXTRA FIELDS FOR COMPOSITED TYPES________________
function onParameterTypeChange(
  parameter:
    | QuestionCodingContentInputDefinition
    | QuestionCodingContentParameter
) {
  if (parameter.type === QuestionCodingContentCompositeInputTypes.ARRAY) {
    if (isNullOrUndefined(parameter.items) || parameter.items.length === 0) {
      parameter.items = [
        { type: 0 as QuestionCodingContentPrimitiveInputTypes },
      ]
    }
  } else {
    parameter.items = undefined
  }
}
</script>
