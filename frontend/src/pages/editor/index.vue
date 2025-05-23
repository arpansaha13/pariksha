<template>
  <main class="h-screen w-screen px-4 py-2">
    <ClientOnly>
      <SplitterGroup
        id="splitter-group-1"
        direction="horizontal"
        auto-save-id="editor-splitter-group-1-save"
      >
        <SplitterPanel
          id="splitter-group-1-panel-1"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard
            :ui="{
              root: 'overflow-hidden h-full',
              body: 'overflow-y-auto h-full',
            }"
          >
            <article class="prose">
              <h1>Lorem Ipsum</h1>
              <p>
                Lorem ipsum dolor sit amet,
                <span>consectetur adipiscing</span> elit. Nullam venenatis justo
                id ante convallis, eget scelerisque tortor ultrices.
              </p>
              <h2>Features</h2>
              <ul>
                <li>Curabitur aliquet quam id dui posuere blandit.</li>
                <li>Pellentesque in ipsum id orci porta dapibus.</li>
                <li>Donec sollicitudin molestie malesuada.</li>
              </ul>
              <h3>Code Example</h3>
              <pre>
  <code>
      function loremIpsum() {
          return "Lorem ipsum dolor sit amet, consectetur adipiscing elit.";
      }
      console.log(loremIpsum());
  </code>
</pre>
              <h2>Conclusion</h2>
              <p>
                Nam ultricies tristique lacus, in vehicula ligula sodales vel.
                <i>
                  Donec sodales leo et arcu interdum, vel vehicula libero
                  volutpat.
                </i>
              </p>
            </article>
          </UCard>
        </SplitterPanel>

        <SplitterResizeHandle
          id="splitter-group-1-resize-handle-1"
          class="group w-2"
        >
          <div
            class="group-hover:bg-primary-500 mx-auto h-full w-0.5 transition-colors delay-150"
          />
        </SplitterResizeHandle>

        <SplitterPanel
          id="splitter-group-1-panel-2"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
            <MonacoEditor
              v-if="editorStore.isEditorPrepared"
              v-model="value"
              :lang="editorLang"
              class="h-full"
            />
          </UCard>
        </SplitterPanel>
      </SplitterGroup>
    </ClientOnly>
  </main>
</template>

<script lang="ts" setup>
definePageMeta({
  layout: 'blank',
})

const editorStore = useEditorStore()

onMounted(async () => {
  await editorStore.prepareEditor()
})

const editorLang = ref<EditorLanguage>('javascript')
const value = ref(`function getColor() {
  const a = ""
}`)
</script>

<style scoped>
@reference "../../assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
