<template>
  <div
    class="relative rounded-sm px-2 py-1 shadow ring-1 ring-black/5 sm:px-2.5 sm:py-1.5"
  >
    <!-- Progress bar background -->
    <div
      class="bg-primary-100 absolute inset-0 rounded-sm transition-all duration-1000"
      :style="{ width: `${remainingPercentage}%` }"
    />

    <!-- Timer text -->
    <div class="relative text-center text-sm font-medium">
      <p class="sm:hidden">
        <span v-if="timer.hrs > 0">{{ timerPadded.hrs }}h</span>
        {{ timerPadded.mins }}m
        <span v-if="timer.hrs === 0">{{ timerPadded.secs }}s</span>
      </p>
      <p class="hidden sm:block">
        <span v-if="timer.hrs > 0">{{ timerPadded.hrs }}h</span>
        {{ timerPadded.mins }}m {{ timerPadded.secs }}s
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  startedAt: string
  scheduledEndTime: string
}>()
const emit = defineEmits(['timeout'])

// Time constants in milliseconds
const MILLISECONDS_PER_SECOND = 1000
const MILLISECONDS_PER_MINUTE = MILLISECONDS_PER_SECOND * 60
const MILLISECONDS_PER_HOUR = MILLISECONDS_PER_MINUTE * 60
const MILLISECONDS_PER_DAY = MILLISECONDS_PER_HOUR * 24

const timer = ref({ hrs: 0, mins: 0, secs: 0 })
const timerPadded = ref({ hrs: '00', mins: '00', secs: '00' })
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

  timer.value.hrs = Math.floor(
    (distance % MILLISECONDS_PER_DAY) / MILLISECONDS_PER_HOUR
  )
  timer.value.mins = Math.floor(
    (distance % MILLISECONDS_PER_HOUR) / MILLISECONDS_PER_MINUTE
  )
  timer.value.secs = Math.floor(
    (distance % MILLISECONDS_PER_MINUTE) / MILLISECONDS_PER_SECOND
  )

  timerPadded.value.hrs = timer.value.hrs.toString().padStart(2, '0')
  timerPadded.value.mins = timer.value.mins.toString().padStart(2, '0')
  timerPadded.value.secs = timer.value.secs.toString().padStart(2, '0')
}

onMounted(() => {
  updateCountdown()
  intervalId = setInterval(updateCountdown, 1000)
})

onUnmounted(() => {
  clearInterval(intervalId)
})
</script>
