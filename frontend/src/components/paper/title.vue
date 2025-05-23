<template>
  <EditableRoot
    v-model="editablePaperTitle"
    name="paper-title"
    activation-mode="focus"
    submit-mode="both"
    placeholder=""
    @submit="updatePaperTitle"
  >
    <EditableArea
      class="rounded-sm px-1 text-xl focus-within:outline hover:outline"
    >
      <EditablePreview as="h1" class="font-semibold" />
      <EditableInput class="font-semibold outline-none" />
    </EditableArea>
  </EditableRoot>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'

const props = defineProps({
  paper: {
    type: Object as PropType<Paper>,
    required: true,
  },
})

const editablePaperTitle = ref(props.paper.title)

function updatePaperTitle() {
  if (isNullOrUndefined(props.paper)) return

  editablePaperTitle.value = editablePaperTitle.value.trim()

  if (!editablePaperTitle.value) {
    editablePaperTitle.value = 'Untitled Paper'
  }
  if (editablePaperTitle.value !== props.paper.title) {
    return updatePaper(props.paper.id, { title: editablePaperTitle.value })
  }
}

watch(props, newProps => {
  editablePaperTitle.value = newProps.paper.title
})
</script>
