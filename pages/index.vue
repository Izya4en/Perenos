<template>
  <div class="page-map">
    
    <div class="map-area">
      <ClientOnly>
        <MapView 
          ref="mapViewRef"
          v-if="!pending && apiData"
          :forte-atms="forteATMs" 
          :competitor-atms="competitorATMs" 
          :heatmap-grid="heatmapGrid"
          :traffic-geo-json="trafficGeoJson"
          :recommendations="aiRecommendations"
          :show-heatmap="showHeatmap"
          :show-traffic="showTraffic"
          :show-recommendations="showRecommendations"
          @select-atm="onMapSelect"
        />
        <div v-else class="loading-overlay">
          <div class="spinner"></div>
          <span>Загрузка карты...</span>
        </div>
      </ClientOnly>
    </div>

    <div class="controls-container left-pos" :class="{ 'shifted': showComplaints }">
      <button class="btn-toggle-complaints" @click="toggleComplaints">
        <span v-if="!showComplaints"> Жалобы Forte</span>
        <span v-else>✕ Закрыть</span>
        <span class="c-badge" v-if="!showComplaints && allComplaints.length > 0">
          {{ allComplaints.length }}
        </span>
      </button>
    </div>

    <div class="controls-container right-pos" :class="{ 'shifted': selectedAtm }">
      <div class="map-controls-group glass-panel">
        <div class="district-selector">
          <label>Район:</label>
          <select @change="e => goToDistrict(astanaDistricts[(e.target as HTMLSelectElement).selectedIndex])">
            <option v-for="(d, idx) in astanaDistricts" :key="idx">{{ d.name }}</option>
          </select>
        </div>
        
        <div class="separator"></div>
        
        <label class="toggle-switch">
          <input type="checkbox" v-model="showHeatmap">
          <span class="slider"></span>
          <span class="label-text">Тран.Траф</span>
        </label>
        
        <label class="toggle-switch">
          <input type="checkbox" v-model="showTraffic">
          <span class="slider"></span>
          <span class="label-text">Траф.Пеш</span>
        </label>

        <label class="toggle-switch">
          <input type="checkbox" v-model="showRecommendations">
          <span class="slider"></span>
          <span class="label-text">Реком. модель</span>
        </label>
        
      </div>
    </div>

    <Transition name="slide-left">
      <aside v-if="showComplaints" class="drawer left-drawer">
        <div class="drawer-header"><h3>Лента жалоб</h3></div>
        <div class="drawer-body">
          <div v-for="(item, idx) in allComplaints" :key="idx" class="feed-item" @click="focusAtmFromComplaint(item.atmObj)">
            <div class="f-top">
              <span class="f-cat">{{ item.c.category || 'Проблема' }}</span>
              <span class="f-date">{{ item.c.date }}</span>
            </div>
            <div class="f-text">{{ item.c.text }}</div>
            <div class="f-link">📍 Показать на карте</div>
          </div>
          <div v-if="!allComplaints.length" class="empty-msg">Нет активных жалоб</div>
        </div>
      </aside>
    </Transition>

    <Transition name="slide-right">
      <aside v-if="selectedAtm" class="drawer right-drawer">
        <div class="drawer-header">
           <div class="top-label"><span class="dot-live">●</span> Live Data</div>
           <div class="bank-label" :class="{ 'not-forte': !selectedAtm.isForte }">
             {{ selectedAtm.isForte ? 'FORTE BANK' : (selectedAtm.bank || 'Конкурент') }}
           </div>
           <h2 class="atm-name">{{ selectedAtm.terminal_id || selectedAtm.name }}</h2>
           <div class="atm-address">{{ selectedAtm.address || 'Адрес не указан' }}</div>
           <div v-if="selectedAtm.isForte" class="status-badge" :class="getBadgeClass(selectedAtm)">
             {{ translateStatus(selectedAtm) }}
           </div>
        </div>

        <div class="drawer-body">
          <div v-if="selectedAtm.isForte">
            
            <div class="stats-row">
               <div class="stat-card">
                 <label>БАЛАНС</label>
                 <div class="val">{{ fmtKZT(selectedAtm.totalCashKZT || 0) }}</div>
               </div>
               <div class="stat-card">
                 <label>ПРОСТОЙ</label>
                 <div class="val">{{ ((selectedAtm.downtimePct || 0) * 100).toFixed(1) }}%</div>
               </div>
            </div>

            <div class="cassettes-section">
              <div class="section-title">Кассеты</div>
              
              <div class="c-group">
                <div class="c-header">
                  <span>Выдача</span>
                  <span class="c-status">OK</span>
                </div>
                <div class="progress-track">
                  <div class="progress-fill blue" :style="{ width: (selectedAtm.dispenseLevel || 45) + '%' }"></div>
                </div>
              </div>

              <div class="c-group">
                <div class="c-header">
                  <span>Прием</span>
                  <span class="c-status">OK</span>
                </div>
                <div class="progress-track">
                  <div class="progress-fill green" :style="{ width: (selectedAtm.depositLevel || 80) + '%' }"></div>
                </div>
              </div>
            </div>
            <div class="section-block complaints-local">
              <h4>Жалобы на терминал</h4>
              <div v-if="selectedAtm.complaints && selectedAtm.complaints.length > 0">
                <div v-for="(c, idx) in selectedAtm.complaints" :key="idx" class="local-complaint">
                  <div class="lc-head">
                    <span class="lc-cat">{{ c.category || 'ERROR' }}</span>
                    <span class="lc-date">{{ c.date }}</span>
                  </div>
                  <div class="lc-text">{{ c.text }}</div>
                </div>
              </div>
              <div v-else class="empty-local"> Жалоб нет</div>
            </div>

            <div class="actions-group">
               <button class="btn-cyan">🛠 Создать заявку</button>
               <button class="btn-outline" @click="selectedAtm = null">Закрыть</button>
            </div>
          </div>

          <div v-else class="competitor-view">
            <div class="section-title">Аналитика конкурента</div>
            <div class="stats-row">
               <div class="stat-card"><label>⬇ Снятие (Est)</label><div class="val text-red">{{ fmtKZT(selectedAtm.estWithdrawalKZT || 0) }}</div></div>
               <div class="stat-card"><label>⬆ Внесение (Est)</label><div class="val text-green">{{ fmtKZT(selectedAtm.estDepositKZT || 0) }}</div></div>
            </div>
            <button class="btn-outline" @click="selectedAtm = null">Закрыть</button>
          </div>
        </div>
      </aside>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// 1. Data Fetching
const { data: apiData, pending } = await useFetch<any>('http://localhost:8080/api/dashboard', { 
  server: false,
  key: 'dashboard-final-merge'
})

// 2. Constants
const astanaDistricts = [
  { name: 'Весь город', lat: 51.147, lng: 71.430, zoom: 11 },
  { name: 'р-н Есиль', lat: 51.128, lng: 71.430, zoom: 14 },
  { name: 'р-н Нура', lat: 51.105, lng: 71.400, zoom: 14 },
  { name: 'р-н Алматы', lat: 51.165, lng: 71.480, zoom: 14 },
  { name: 'р-н Сарыарка', lat: 51.185, lng: 71.415, zoom: 14 },
  { name: 'р-н Байконур', lat: 51.160, lng: 71.450, zoom: 14 },
]
const selectedDistrict = ref(astanaDistricts[0])

// 3. Logic & Computed
const allAtms = computed(() => {
  const d = apiData.value
  let list: any[] = []
  if (d && (d.forte || d.competitors)) list = [...(d.forte || []), ...(d.competitors || [])]
  else if (Array.isArray(d)) list = d

  return list.map(item => {
    const isForte = (item.bank || '').toLowerCase().includes('forte')
    let complaints = item.complaints
    try { if (typeof complaints === 'string') complaints = JSON.parse(complaints) } catch {}
    
    // Normalize Data
    const tId = item.terminal_id || item.name 
    const lat = Number(item.lat || item.latitude)
    const lng = Number(item.lng || item.longitude || item.lon)

    return { 
      ...item, 
      isForte, 
      terminal_id: tId, 
      complaints: Array.isArray(complaints) ? complaints : [],
      lat, lng
    }
  })
})

const forteATMs = computed(() => allAtms.value.filter(a => a.isForte))
const competitorATMs = computed(() => allAtms.value.filter(a => !a.isForte))
const heatmapGrid = computed(() => apiData.value?.heatmapGrid || null)
const trafficGeoJson = computed(() => {
  if (!apiData.value?.traffic) return null
  return {
    type: 'FeatureCollection',
    features: apiData.value.traffic.map((item: any) => ({
      type: 'Feature',
      geometry: typeof item.geometry === 'string' ? JSON.parse(item.geometry) : item.geometry,
      properties: { traffic: item.weekday_traffic }
    }))
  }
})

// === AI Mock Data ===
const aiRecommendations = ref([
  { lat: 51.135, lng: 71.445, score: 94, predictedTurnover: 25000000, reason: "Высокий трафик" },
  { lat: 51.155, lng: 71.425, score: 82, predictedTurnover: 18000000, reason: "Новый ЖК" }
])

// 4. UI State
const showHeatmap = ref(true)
const showTraffic = ref(true)
const showRecommendations = ref(false)
const showComplaints = ref(false)
const selectedAtm = ref<any>(null)
const mapViewRef = ref<any>(null)

const allComplaints = computed(() => {
  const list: any[] = []
  forteATMs.value.forEach(atm => {
    atm.complaints.forEach((c: any) => {
      list.push({ c, atmName: atm.terminal_id || atm.name, atmObj: atm })
    })
  })
  return list.sort((a, b) => new Date(b.c.date).getTime() - new Date(a.c.date).getTime())
})

// 5. Methods
function toggleComplaints() {
  showComplaints.value = !showComplaints.value
  if(showComplaints.value) selectedAtm.value = null
}

function onMapSelect(atm: any) { 
  if (!atm) { selectedAtm.value = null; return }
  
  // Поиск обогащенного объекта
  let enriched = allAtms.value.find(a => 
    String(a.terminal_id) === String(atm.terminal_id) ||
    String(a.id) === String(atm.id)
  )
  
  // Фолбэк по координатам
  if (!enriched) {
    enriched = allAtms.value.find(a => 
      Math.abs(Number(a.lat) - Number(atm.lat)) < 0.00001 && 
      Math.abs(Number(a.lng) - Number(atm.lng)) < 0.00001
    )
  }
  
  selectedAtm.value = enriched || atm
  showComplaints.value = false 
  
  if (mapViewRef.value) mapViewRef.value.flyToLocation(atm.lat, atm.lng, 16)
}

function focusAtmFromComplaint(atm: any) {
  showComplaints.value = false
  onMapSelect(atm)
}

function goToDistrict(district: any) {
  selectedDistrict.value = district
  if (mapViewRef.value) mapViewRef.value.flyToLocation(district.lat, district.lng, district.zoom)
}

function fmtKZT(n: any) { return new Intl.NumberFormat('ru-KZ',{style:'currency',currency:'KZT',maximumFractionDigits:0}).format(n || 0) }
function getBadgeClass(atm: any) { return (atm.efficiencyStatus || '').toLowerCase().includes('ineffective') ? 'Ineffective' : 'Effective' }
function translateStatus(atm: any) { return getBadgeClass(atm) === 'Effective' ? 'Эффективный' : 'Проблемный' }

// 6. Lifecycle (URL Handling)
onMounted(async () => {
  const targetId = route.query.atmId
  
  if (targetId && apiData.value) {
    await nextTick()
    const target = allAtms.value.find(a => 
      String(a.terminal_id) === String(targetId) || 
      String(a.id) === String(targetId)
    )
    if (target) onMapSelect(target)
  }
})
</script>

<style scoped>
/* GENERAL LAYOUT */
.page-map { display: flex; flex-direction: column; height: 100vh; overflow: hidden; position: relative; background: #0f172a; }
.map-area { flex: 1; position: relative; }
.loading-overlay { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: #fff; background: #0f172a; }

/* CONTROLS */
.controls-container { position: absolute; top: 20px; z-index: 20; transition: transform 0.3s; }
.left-pos { left: 20px; }
.right-pos { right: 20px; }
.left-pos.shifted { transform: translateX(360px); }
.right-pos.shifted { transform: translateX(-380px); }

/* TOGGLES & BUTTONS */
.btn-toggle-complaints { background: #f59e0b; color: #000; border: none; padding: 12px 20px; border-radius: 30px; font-weight: bold; cursor: pointer; display: flex; align-items: center; gap: 10px; box-shadow: 0 4px 15px rgba(245, 158, 11, 0.3); transition: 0.2s; }
.btn-toggle-complaints:hover { transform: scale(1.05); }
.c-badge { background: #000; color: #f59e0b; padding: 2px 8px; border-radius: 12px; font-size: 12px; }

.glass-panel { background: rgba(15, 23, 42, 0.9); backdrop-filter: blur(8px); border: 1px solid #334155; padding: 12px; border-radius: 12px; display: flex; flex-direction: column; gap: 10px; min-width: 180px; }
.district-selector { display: flex; flex-direction: column; gap: 5px; }
.district-selector label { font-size: 11px; color: #94a3b8; font-weight: 600; }
.district-selector select { background: #1e293b; color: #fff; border: 1px solid #334155; padding: 6px; border-radius: 6px; outline: none; cursor: pointer; font-size: 13px; font-weight: bold; }
.separator { height: 1px; background: #334155; margin: 5px 0; }

.toggle-switch { display: flex; align-items: center; cursor: pointer; }
.toggle-switch input { display: none; }
.slider { width: 36px; height: 20px; background-color: #334155; border-radius: 20px; position: relative; margin-right: 8px; transition: .3s; }
.slider:before { content: ""; position: absolute; height: 16px; width: 16px; left: 2px; bottom: 2px; background-color: white; border-radius: 50%; transition: .3s; }
input:checked + .slider { background-color: #00B8D9; }
input:checked + .slider:before { transform: translateX(16px); }
.label-text { font-size: 13px; color: #fff; font-weight: 600; }

/* DRAWERS */
.drawer { position: absolute; top: 0; bottom: 0; width: 360px; background: #0f172a; z-index: 30; display: flex; flex-direction: column; box-shadow: 0 0 40px rgba(0,0,0,0.6); border-color: #1e293b; border-style: solid; }
.left-drawer { left: 0; border-right-width: 1px; }
.right-drawer { right: 0; border-left-width: 1px; }
.drawer-header { padding: 25px; border-bottom: 1px solid #1e293b; background: #162032; }
.drawer-body { flex: 1; overflow-y: auto; padding: 20px; }

/* DRAWER CONTENT */
.atm-name { color: #fff; font-size: 20px; margin: 10px 0 2px 0; }
.atm-address { color: #94a3b8; font-size: 13px; margin-bottom: 10px; }
.bank-label { background: #0e4f5f; color: #00B8D9; padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: bold; display: inline-block; }
.bank-label.not-forte { background: #334155; color: #cbd5e1; }
.status-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; }
.status-badge.Effective { background: rgba(16, 185, 129, 0.2); color: #10b981; }
.status-badge.Ineffective { background: rgba(239, 68, 68, 0.2); color: #ef4444; }

.stats-row { display: flex; gap: 10px; margin-bottom: 20px; }
.stat-card { flex: 1; background: #1e293b; padding: 15px; border-radius: 8px; text-align: center; border: 1px solid #334155; }
.stat-card label { font-size: 10px; color: #94a3b8; display: block; margin-bottom: 5px; }
.stat-card .val { font-size: 16px; font-weight: bold; color: #fff; }

/* === CASSETTES (NEW) === */
.cassettes-section { margin-top: 20px; margin-bottom: 25px; border-top: 1px solid #334155; padding-top: 15px; }
.section-title { color: #94a3b8; font-size: 12px; font-weight: 600; margin-bottom: 12px; text-transform: uppercase; }
.c-group { margin-bottom: 15px; }
.c-header { display: flex; justify-content: space-between; color: #cbd5e1; font-size: 13px; margin-bottom: 6px; }
.c-status { color: #94a3b8; font-size: 11px; }
.progress-track { height: 6px; background: #334155; border-radius: 3px; overflow: hidden; }
.progress-fill { height: 100%; border-radius: 3px; transition: width 0.5s ease; }
.progress-fill.blue { background-color: #3b82f6; box-shadow: 0 0 10px rgba(59, 130, 246, 0.4); }
.progress-fill.green { background-color: #10b981; box-shadow: 0 0 10px rgba(16, 185, 129, 0.4); }

/* COMPLAINTS & FEED */
.feed-item { background: rgba(255,255,255,0.04); padding: 15px; border-radius: 8px; margin-bottom: 10px; cursor: pointer; border-left: 3px solid #f59e0b; }
.f-top { display: flex; justify-content: space-between; font-size: 11px; margin-bottom: 5px; }
.f-cat { color: #f59e0b; font-weight: bold; }
.f-date { color: #64748b; }
.f-text { color: #e2e8f0; font-size: 13px; margin-bottom: 8px; }
.f-link { color: #00B8D9; font-size: 12px; text-decoration: underline; text-align: right; }

.complaints-local { margin-top: 20px; border-top: 1px solid #334155; padding-top: 15px; }
.complaints-local h4 { color: #94a3b8; font-size: 12px; text-transform: uppercase; margin-bottom: 10px; }
.local-complaint { background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); padding: 10px; border-radius: 6px; margin-bottom: 8px; }
.lc-head { display: flex; justify-content: space-between; font-size: 11px; margin-bottom: 4px; }
.lc-cat { color: #ef4444; font-weight: bold; }
.lc-date { color: #cbd5e1; }
.lc-text { color: #fff; font-size: 13px; line-height: 1.4; }
.empty-local { text-align: center; padding: 10px; color: #10b981; font-size: 13px; background: rgba(16, 185, 129, 0.1); border-radius: 6px; }

/* ACTIONS */
.actions-group { display: flex; flex-direction: column; gap: 10px; }
.btn-cyan { background: #00B8D9; border: none; padding: 12px; border-radius: 6px; width: 100%; font-weight: bold; cursor: pointer; margin-bottom: 10px; color: #fff; font-size: 14px; transition: 0.2s; }
.btn-cyan:hover { opacity: 0.9; }
.btn-outline { background: transparent; border: 1px solid #334155; padding: 12px; border-radius: 6px; width: 100%; color: #94a3b8; cursor: pointer; font-size: 14px; transition: 0.2s; }
.btn-outline:hover { border-color: #94a3b8; color: #fff; }

.text-red { color: #ef4444; }
.text-green { color: #10b981; }

.slide-left-enter-active, .slide-left-leave-active, .slide-right-enter-active, .slide-right-leave-active { transition: transform 0.3s; }
.slide-left-enter-from, .slide-left-leave-to { transform: translateX(-100%); }
.slide-right-enter-from, .slide-right-leave-to { transform: translateX(100%); }
</style>