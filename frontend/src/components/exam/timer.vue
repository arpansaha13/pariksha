<template>
  <div class="relative rounded-sm px-2.5 py-1.5 shadow ring-1 ring-black/5">
    <!-- Progress bar background -->
    <div
      class="bg-primary-100 absolute inset-0 rounded-sm transition-all duration-1000"
      :style="{ width: `${remainingPercentage}%` }"
    />
    <!-- Timer text -->
    <p class="relative w-[8ch] text-center text-sm font-medium">
      {{ hours }}:{{ minutes }}:{{ seconds }}
    </p>
  </div>
</template>

<script setup lang="ts">
const props = defineProps({
  startedAt: {
    type: String,
    required: true,
  },
  scheduledEndTime: {
    type: String,
    required: true,
  },
})
const emit = defineEmits(['timeout'])

// Time constants in milliseconds
const MILLISECONDS_PER_SECOND = 1000
const MILLISECONDS_PER_MINUTE = MILLISECONDS_PER_SECOND * 60
const MILLISECONDS_PER_HOUR = MILLISECONDS_PER_MINUTE * 60
const MILLISECONDS_PER_DAY = MILLISECONDS_PER_HOUR * 24

const hours = ref('00')
const minutes = ref('00')
const seconds = ref('00')
const endTime = new Date(props.scheduledEndTime).getTime()
const startTime = new Date(props.startedAt).getTime()
const remainingPercentage = ref(100)
let intervalId: NodeJS.Timeout

const updateCountdown = () => {
  const now = new Date().getTime()
  const distance = endTime - now
  const totalDuration = endTime - startTime

  if (distance <= 0) {
    clearInterval(intervalId as unknown as number)
    remainingPercentage.value = 0
    emit('timeout')
    return
  }

  // Calculate remaining percentage
  remainingPercentage.value = (distance / totalDuration) * 100

  const hrs = Math.floor(
    (distance % MILLISECONDS_PER_DAY) / MILLISECONDS_PER_HOUR
  )
  const mins = Math.floor(
    (distance % MILLISECONDS_PER_HOUR) / MILLISECONDS_PER_MINUTE
  )
  const secs = Math.floor(
    (distance % MILLISECONDS_PER_MINUTE) / MILLISECONDS_PER_SECOND
  )

  hours.value = hrs.toString().padStart(2, '0')
  minutes.value = mins.toString().padStart(2, '0')
  seconds.value = secs.toString().padStart(2, '0')
}

onMounted(() => {
  updateCountdown()
  intervalId = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  clearInterval(intervalId)
})
</script>
