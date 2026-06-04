<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import TopBar from './TopBar.vue'
</script>

<template>
  <div
    class="font-body-md text-body-md antialiased overflow-hidden flex h-screen w-screen selection:bg-primary-container selection:text-on-primary-container"
  >
    <Sidebar />
    <!-- Main Content Wrapper -->
    <main class="flex-1 flex flex-col md:ml-72 relative h-full">
      <TopBar />
      <!-- Routed content area -->
      <div class="flex-1 mt-16 h-[calc(100vh-4rem)] overflow-y-auto">
        <slot />
      </div>
    </main>
  </div>
</template>
