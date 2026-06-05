<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script>
export default { name: 'App' }
</script>

<style>
@import './assets/theme.css';

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  font-family: var(--ds-font);
  color: var(--ds-text);
  background: var(--ds-bg-muted);
  -webkit-font-smoothing: antialiased;
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-thumb {
  background: #d0d0d0;
  border-radius: 3px;
}
::-webkit-scrollbar-track {
  background: transparent;
}

.el-message {
  border-radius: var(--ds-radius) !important;
}
</style>
