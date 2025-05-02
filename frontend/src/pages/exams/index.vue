<template>
  <main>
    <UCard v-if="!isNullOrUndefined(examsTableData)" :ui="{ body: '!py-2' }">
      <template #header>
        <h1 class="heading">Exams</h1>
      </template>

      <UTable :data="examsTableData" :columns="examsTableColumns" />
    </UCard>
  </main>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'

const { data: examsTableData, columns: examsTableColumns } =
  await useExamsTableData()
</script>
