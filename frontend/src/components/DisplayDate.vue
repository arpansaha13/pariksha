<template>
  <ClientOnly>
    <p :class="mergedUi.base">
      <time>
        {{ df.format(date.toDate(getLocalTimeZone())) }}
      </time>
    </p>

    <template #fallback>
      <USkeleton :class="mergedUi.skeleton" />
    </template>
  </ClientOnly>
</template>

<script setup lang="ts">
import {
  CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'

interface DisplayDateUI {
  base: string
  skeleton: string
}

const defaultUi: DisplayDateUI = {
  base: '',
  skeleton: '',
}

const props = defineProps({
  date: {
    type: CalendarDateTime,
    required: true,
  },
  df: {
    type: DateFormatter,
    required: true,
  },
  ui: {
    type: Object as PropType<Partial<DisplayDateUI>>,
    default: () => ({
      base: '',
      skeleton: '',
    }),
  },
})

const mergedUi = computed(() => ({ ...defaultUi, ...props.ui }))
</script>
