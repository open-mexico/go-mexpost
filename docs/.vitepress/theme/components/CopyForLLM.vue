<script setup lang="ts">
import { ref } from "vue";

type CopyStatus = "idle" | "copied" | "error";

const props = withDefaults(
  defineProps<{
    title?: string;
    content: string;
    language?: string;
    buttonText?: string;
  }>(),
  {
    title: "LLM Prompt",
    language: "text",
    buttonText: "Copy for LLM",
  },
);

const status = ref<CopyStatus>("idle");

const copyToClipboard = async (): Promise<void> => {
  try {
    await navigator.clipboard.writeText(props.content.trim());
    status.value = "copied";
    window.setTimeout(() => {
      status.value = "idle";
    }, 1800);
  } catch {
    status.value = "error";
  }
};
</script>

<template>
  <section class="llm-copy-block">
    <header class="llm-copy-header">
      <div>
        <strong>{{ title }}</strong>
      </div>
      <button type="button" class="llm-copy-button" @click="copyToClipboard">
        <span v-if="status === 'idle'">{{ buttonText }}</span>
        <span v-else-if="status === 'copied'">Copied</span>
        <span v-else>Copy failed</span>
      </button>
    </header>

    <pre class="llm-copy-pre"><code>{{ content }}</code></pre>
    <p class="llm-copy-hint">Language: {{ language }}</p>
  </section>
</template>
