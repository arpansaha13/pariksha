<template>
  <main class="flex h-screen w-screen flex-col gap-2 px-4 py-2">
    <slot
      name="header"
      :panelIsCollapsed="panelRef?.isCollapsed ?? false"
      :togglePanel="toggleSplitterPanel"
    />

    <SplitterGroup
      :id="Splitter.GROUP_1_ID"
      direction="horizontal"
      @layout="splitterGroup1Layout = $event"
    >
      <SplitterPanel
        :id="Splitter.PRIMARY_SIDE_BAR"
        :default-size="splitterGroup1Layout?.[0]"
        :min-size="20"
        class="splitter-panel"
      >
        <UCard
          :ui="{
            root: 'overflow-hidden h-full',
            body: 'overflow-y-auto h-full',
          }"
        >
          <slot name="primary-sidebar" />
        </UCard>
      </SplitterPanel>

      <SplitterResizeHandle
        :id="Splitter.GROUP_1_RESIZE_HANDLE_1_ID"
        class="group w-2 outline-none"
      >
        <div class="splitter-resizer-handle mx-auto h-full w-0.5">
          <div class="splitter-resizer-marker h-4 w-[2px]" />
        </div>
      </SplitterResizeHandle>

      <SplitterPanel
        :id="Splitter.GROUP_1_PANEL_2_ID"
        :default-size="splitterGroup1Layout?.[1]"
        :min-size="20"
      >
        <SplitterGroup
          :id="Splitter.GROUP_2_ID"
          ref="splitterGroup2Ref"
          direction="vertical"
          @layout="splitterGroup2Layout = $event"
        >
          <SplitterPanel
            :id="Splitter.GROUP_2_PANEL_1_ID"
            :default-size="splitterGroup2Layout?.[0]"
            :min-size="20"
            class="splitter-panel"
          >
            <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
              <slot />
            </UCard>
          </SplitterPanel>

          <SplitterResizeHandle
            :id="Splitter.GROUP_2_RESIZE_HANDLE_1_ID"
            class="group flex h-2 flex-col outline-none"
          >
            <div class="splitter-resizer-handle my-auto h-0.5 w-full">
              <div class="splitter-resizer-marker h-[2px] w-4" />
            </div>
          </SplitterResizeHandle>

          <SplitterPanel
            :id="Splitter.GROUP_2_PANEL_2_ID"
            ref="panelRef"
            collapsible
            :default-size="splitterGroup2Layout?.[1]"
            :min-size="splitterGroup2Panel2CollapsedSize"
            :collapsed-size="splitterGroup2Panel2CollapsedSize"
            class="splitter-panel"
          >
            <UCard
              :ui="{
                root: 'h-full flex flex-col',
                body: 'h-full p-0!',
              }"
            >
              <slot
                name="panel"
                :toggle="toggleSplitterPanel"
                :isCollapsed="panelRef?.isCollapsed"
              />
            </UCard>
          </SplitterPanel>
        </SplitterGroup>
      </SplitterPanel>
    </SplitterGroup>
  </main>
</template>

<script lang="ts" setup>
import type { SplitterGroup, SplitterPanel } from '#components'
import { isNullOrUndefined } from '@arpansaha13/utils'

enum Splitter {
  GROUP_1_ID = 'splitter-group-1',
  PRIMARY_SIDE_BAR = 'primary-side-bar', // splitter-group-1-panel-1
  GROUP_1_PANEL_2_ID = 'splitter-group-1-panel-2',
  GROUP_1_RESIZE_HANDLE_1_ID = 'splitter-group-1-resize-handle-1',

  GROUP_2_ID = 'splitter-group-2',
  GROUP_2_PANEL_1_ID = 'splitter-group-2-panel-1',
  GROUP_2_PANEL_2_ID = 'splitter-group-2-panel-2',
  GROUP_2_RESIZE_HANDLE_1_ID = 'splitter-group-2-resize-handle-1',
}

// ______________________SPLITTER LAYOUT SSR______________________
const splitterGroup1Layout = useCookie<number[]>(
  'editor:splitter-group-1-layout',
  { path: '/editor' }
)
const splitterGroup2Layout = useCookie<number[]>(
  'editor:splitter-group-2-layout',
  { path: '/editor' }
)

// ____________CALCULATE GROUP-2 PANEL-2 COLLAPSED SIZE___________
const splitterGroup2Ref = ref<InstanceType<typeof SplitterGroup>>()
const panelRef = ref<InstanceType<typeof SplitterPanel>>()

const splitterGroup2Panel2CollapsedSize = computed(() => {
  if (!splitterGroup2Ref.value) return 0

  const el = splitterGroup2Ref.value.$el as HTMLDivElement
  const splitterGroup2Height = el.clientHeight
  const splitterGroup2Panel2HeaderHeight = 36.8 // Update this if header height changes
  const splitterGroup2ResizeHandlerHeight = 0 // Update this if resize-handler height changes
  const effectiveAvailableHeight =
    splitterGroup2Panel2HeaderHeight - splitterGroup2ResizeHandlerHeight
  const collapsedSizePercent =
    (effectiveAvailableHeight / splitterGroup2Height) * 100

  return Math.floor(collapsedSizePercent * 100) / 100 // 2 decimal places
})

// ________________SPLITTER PANEL EXPAND/COLLAPSE_________________
function useSplitterToggleCollapse() {
  const isPanelInitiallyCollapsed = ref({} as Record<Splitter, boolean>)

  const { stop } = watchEffect(async () => {
    if (panelRef.value?.isCollapsed) {
      isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID] = true
      stop()
    }
  })

  function toggleSplitterPanel() {
    const panelInstance = panelRef.value
    if (isNullOrUndefined(panelInstance)) return

    if (panelInstance.isExpanded) {
      panelInstance.collapse()
      return
    }

    if (isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID]) {
      panelInstance.resize(45)
      isPanelInitiallyCollapsed.value[Splitter.GROUP_2_PANEL_2_ID] = false
    } else {
      panelInstance.expand()
    }
  }

  return toggleSplitterPanel
}

const toggleSplitterPanel = useSplitterToggleCollapse()
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}

.splitter-resizer-handle {
  @apply group-hover:bg-primary-500 group-focus:bg-primary-500 flex items-center justify-center transition-colors delay-150;
}

.splitter-resizer-marker {
  @apply rounded-full bg-neutral-400 transition-colors delay-150 group-hover:bg-transparent group-focus:bg-transparent;
}
</style>
