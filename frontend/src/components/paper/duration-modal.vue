<template>
  <UModal
    title="Edit duration"
    description="Edit the paper duration"
    @after:leave="updatePaperDuration"
  >
    <UTooltip text="Edit duration">
      <UButton
        icon="i-lucide-alarm-clock"
        size="sm"
        color="neutral"
        variant="outline"
        class="ml-auto"
      />
    </UTooltip>

    <template #body>
      <div class="grid grid-cols-2 gap-x-4">
        <UFormField label="Hours" name="duration_hours">
          <UInputNumber v-model="hours" :min="0" :max="24" />
        </UFormField>

        <UFormField label="Minutes" name="duration_minutes">
          <UInputNumber v-model="minutes" :min="0" :max="59" />
        </UFormField>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import type { Paper } from '~/types/paper'

const props = defineProps({
  paper: {
    type: Object as PropType<Paper>,
    required: true,
  },
})

const hours = ref(0)
const minutes = ref(0)

watchEffect(() => {
  if (props.paper.duration_minutes) {
    hours.value = Math.floor(props.paper.duration_minutes / 60)
    minutes.value = props.paper.duration_minutes % 60
  }
})

async function updatePaperDuration() {
  const totalMinutes = hours.value * 60 + minutes.value
  if (totalMinutes > 0) {
    await updatePaper(props.paper.id, { duration_minutes: totalMinutes })
  }
}
</script>
