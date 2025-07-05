<template>
  <USlideover
    side="left"
    v-model:open="isOpen"
    :ui="{ title: 'flex items-center gap-1.5', description: 'm-0!' }"
  >
    <UContainer class="bg-default py-2.5 shadow-sm sm:hidden">
      <UButton icon="heroicons:bars-3" color="neutral" variant="subtle" />
    </UContainer>

    <template #title>
      <Logo class="size-8" />
      <p class="text-md font-bold">Pariksha</p>
    </template>

    <template #description>
      <p class="sr-only">Navbar</p>
    </template>

    <template #body>
      <Sidebar />
    </template>
  </USlideover>

  <UContainer class="min-h-screen gap-6 sm:flex">
    <div class="sticky top-0 hidden h-screen w-60 shrink-0 sm:block">
      <Sidebar class="p-2">
        <template #leading>
          <div class="flex items-center gap-1.5 p-2">
            <Logo class="size-8" />
            <p class="text-md font-bold">Pariksha</p>
          </div>
        </template>
      </Sidebar>
    </div>

    <div class="grow py-4">
      <slot />
    </div>
  </UContainer>
</template>

<script setup lang="ts">
const isOpen = ref(false)

const route = useRoute()
watch(
  () => route.path,
  (newVal, oldVal) => {
    if (newVal !== oldVal) {
      isOpen.value = false
    }
  }
)
</script>
