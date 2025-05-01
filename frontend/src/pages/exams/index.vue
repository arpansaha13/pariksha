<template>
  <main>
    <h1 class="heading mb-6">Exams</h1>

    <UCard v-if="examsTableData !== null" :ui="{ body: 'p-0 sm:p-0' }">
      <UTable :data="examsTableData" :columns="examsTableColumns" />
    </UCard>
  </main>
</template>

<script setup lang="ts">
const { data: examsTableData, columns: examsTableColumns } =
  await useExamsTableData()
</script>
