<script setup lang="ts">
definePageMeta({
  layout: 'blank',
  middleware: async (_to, from) => {
    const currentPath = '/papers/new'

    // Check if the current url was directly entered in address bar
    if (from.path === currentPath) {
      return navigateTo('/', { replace: true })
    }

    const newPaper = await createPaper()
    return navigateTo(`/papers/${newPaper.id}`, { replace: true })
  },
})
</script>
