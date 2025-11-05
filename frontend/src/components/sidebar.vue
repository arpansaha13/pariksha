<template>
  <div class="flex h-full shrink-0 flex-col">
    <slot name="leading" />

    <UNavigationMenu
      :class="$slots.leading && 'mt-2'"
      :items="links"
      orientation="vertical"
    />

    <UDropdownMenu
      v-if="authUser"
      :items="profileMenuItems"
      :content="{
        align: 'center',
        side: 'top',
        sideOffset: 8,
      }"
      :ui="{
        content: 'w-48',
      }"
    >
      <UButton
        color="neutral"
        variant="ghost"
        class="mt-auto flex items-center gap-x-1.5 p-2 text-left"
      >
        <UUser
          size="lg"
          :name="
            authUser.first_name
              ? authUser.first_name + ' ' + authUser.last_name
              : authUser.email
          "
          :description="`@${authUser.username}`"
          :avatar="{ icon: 'heroicons:user-20-solid' }"
        />
      </UButton>
    </UDropdownMenu>
  </div>
</template>

<script setup lang="ts">
import type { DropdownMenuItem, NavigationMenuItem } from '@nuxt/ui'

const { data: authUser } = await useAuthUser()

const profileMenuItems = ref<DropdownMenuItem[]>([
  {
    label: 'Logout',
    icon: 'lucide:log-out',
    onSelect: async () => {
      await logout()
      reloadNuxtApp({ persistState: false, path: '/auth/login' })
    },
  },
])

const links: NavigationMenuItem[][] = [
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
      label: 'Create new exam',
      to: '/exams/new',
      icon: 'i-lucide-bookmark-plus',
    },
    {
      label: 'Create new paper',
      to: '/papers/new',
      icon: 'i-heroicons-document-plus',
    },
  ],
]
</script>
