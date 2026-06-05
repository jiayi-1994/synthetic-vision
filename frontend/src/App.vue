<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import { useCursorGlow } from './composables/useInteractions'

const route = useRoute()
const isPublic = computed(() => route.meta.public === true)

// Global cursor light (hover devices only; no-op on touch / reduced-motion).
useCursorGlow()
</script>

<template>
  <!-- Ambient neon backdrop + perspective grid floor, behind everything -->
  <div class="app-backdrop"></div>
  <div class="grid-floor"></div>
  <div class="cursor-glow"></div>

  <RouterView v-if="isPublic" />
  <AppShell v-else>
    <RouterView />
  </AppShell>
</template>
