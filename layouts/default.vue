<template>
  <div class="layout-container">
    <header class="main-header">
      <div class="brand">
        <strong>Forte GeoAnalytics</strong>
        <span class="tag">ATM</span>
      </div>
      
      <nav class="nav-links">
        <NuxtLink to="/" class="nav-item" active-class="active">
          🌍 Карта
        </NuxtLink>
        <NuxtLink to="/analytics" class="nav-item" active-class="active">
          📊 Графики
        </NuxtLink>
      </nav>

      <div class="user">admin@forte.kz</div>
    </header>

    <div class="content">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.layout-container { display: flex; flex-direction: column; height: 100vh; background: #020617; color: #e2e8f0; font-family: 'Segoe UI', sans-serif; }
.main-header { height: 55px; background: #0f172a; border-bottom: 1px solid #1e293b; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; }
.brand strong { font-size: 18px; color: #fff; margin-right: 8px; }
.tag { font-size: 11px; background: #00B8D9; color: #000; padding: 2px 6px; border-radius: 4px; font-weight: bold; }
.nav-links { display: flex; gap: 20px; height: 100%; }
.nav-item { display: flex; align-items: center; color: #94a3b8; text-decoration: none; padding: 0 10px; height: 100%; border-bottom: 2px solid transparent; transition: 0.2s; cursor: pointer; }
.nav-item:hover { color: #fff; background: rgba(255,255,255,0.03); }
.nav-item.active { color: #00B8D9; border-bottom-color: #00B8D9; font-weight: 600; }
.content { flex: 1; overflow: hidden; position: relative; }
</style>