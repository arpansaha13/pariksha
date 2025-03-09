<template>
  <UContainer class="flex h-screen gap-6">
    <div class="w-60 p-2">
      <div class="flex items-center gap-2 p-2">
        <Logo class="size-8" />
        <p class="text-md font-bold">Pariksha</p>
      </div>

      <UVerticalNavigation
        class="mt-2"
        :links="links"
        :ui="{
          padding: 'px-2.5 py-2',
        }"
      />
    </div>

    <div class="flex-grow py-4">
      <slot />
    </div>
  </UContainer>
</template>

<script setup lang="ts">
const links = [
  [
    {
      label: 'Home',
      to: '/',
      icon: 'i-heroicons-home',
    },
    {
      label: 'Exams',
      to: '/exams',
      icon: 'i-heroicons-academic-cap',
    },
    {
      label: 'Papers',
      to: '/papers',
      icon: 'i-heroicons-document-text',
    },
  ],
  [
    {
      label: 'Create new paper',
      to: '/papers/new',
      icon: 'i-heroicons-document-plus',
    },
  ],
]
</script>
