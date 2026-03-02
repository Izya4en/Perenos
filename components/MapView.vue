<template>
  <div ref="mapEl" class="map"></div>

  <div v-if="hoverInfo" class="tooltip" :style="{ left: hoverInfo.x + 'px', top: hoverInfo.y + 'px' }">
    
    <div v-if="hoverInfo.type === 'atm'">
      <div class="name">{{ hoverInfo.props.name || hoverInfo.props.terminal_id }}</div>
      
      <div class="row header-row">
        <span>Банк:</span> 
        <b :style="{color: hoverInfo.props.isForte ? '#00B8D9' : '#9E9E9E'}">
          {{ hoverInfo.props.isForte ? 'FORTE BANK' : (hoverInfo.props.bank || 'Конкурент') }}
        </b>
      </div>

      <div v-if="hoverInfo.props.isForte" class="forte-data">
        <div class="sep"></div>
        <div class="row">
          <span>Остаток:</span> 
          <b>{{ fmtKZT(hoverInfo.props.totalCashKZT) }}</b>
        </div>
        <div class="row">
           <span>Статус:</span> 
           <b :style="{ color: getStatusColor(hoverInfo.props.efficiencyStatus) }">
             {{ translateStatus(hoverInfo.props.efficiencyStatus) }}
           </b>
        </div>
      </div>
      
      <div v-else class="comp-data">
        <div class="sep"></div>
        <div class="row flow-row">
          <span style="color: #ef4444;">⬇ Снятие (Est):</span> 
          <b>{{ fmtKZT(hoverInfo.props.estWithdrawalKZT) }}</b>
        </div>
      </div>
    </div>

    <div v-else-if="hoverInfo.type === 'recommendation'">
      <div class="name" style="color: #c084fc">✨ Рекомендация ИИ</div>
      <div class="sep"></div>
      <div class="row"><span>Прогноз оборота:</span> <b>{{ fmtKZT(hoverInfo.props.predictedTurnover) }}</b></div>
      <div class="row"><span>Рейтинг (Score):</span> <b style="color: #10b981">{{ hoverInfo.props.score }}/100</b></div>
      <div class="row" style="margin-top: 5px;">
        <span style="display:block; font-size: 11px; color: #94a3b8;">{{ hoverInfo.props.reason }}</span>
      </div>
    </div>

    <div v-else-if="hoverInfo.type === 'traffic'">
      <div class="name">🛣 П.трафик</div>
      <div class="sep"></div>
      <div class="row">
        <span>Чел/сутки:</span>
        <b :style="{ color: getTrafficColor(hoverInfo.props.traffic) }">
          {{ hoverInfo.props.traffic }}
        </b>
      </div>
    </div>

    <div v-else-if="hoverInfo.type === 'grid'">
      <div class="row">Популярность:</div>
      <div class="val" :style="{ color: hoverInfo.color }">
        {{ (hoverInfo.weight * 100).toFixed(0) }}%
      </div>
    </div>

    <div v-else-if="hoverInfo.type === 'district'">
      <div class="name" style="color: #fbbf24">📍 Район {{ hoverInfo.props.shapeName }}</div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch, computed } from 'vue'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'

const props = defineProps<{
  forteAtms: any[]
  competitorAtms: any[]
  heatmapGrid: any       
  trafficGeoJson: any      
  showHeatmap: boolean
  showTraffic: boolean
  // ДОБАВЛЕНЫ ПРОПСЫ ДЛЯ ИИ
  recommendations?: any[]
  showRecommendations?: boolean
}>()

const emit = defineEmits(['select-atm'])

const mapEl = ref<HTMLDivElement | null>(null)
let map: maplibregl.Map | null = null
const hoverInfo = ref<any>(null)

// Форматтеры
function fmtKZT(n: number) { return new Intl.NumberFormat('ru-KZ', { style: 'currency', currency: 'KZT', maximumFractionDigits: 0 }).format(n || 0) }
function getStatusColor(s: string) { return s === 'Effective' ? '#10b981' : s === 'Ineffective' ? '#ef4444' : '#00B8D9' }
function translateStatus(s: string) { return s === 'Effective' ? 'Эффективный' : s === 'Ineffective' ? 'Неэффективный' : 'Норма' }
function getTrafficColor(val: number) { return val > 1500 ? '#ef4444' : val > 500 ? '#fcd34d' : '#10b981' }

// Метод для перемещения камеры
function flyToLocation(lat: number, lng: number, zoom: number = 16) {
  if (map) {
    map.flyTo({ 
      center: [lng, lat], 
      zoom: zoom, 
      speed: 1.5, 
      essential: true 
    })
  }
}
defineExpose({ flyToLocation })

// Подготовка данных
const forteFC = computed(() => ({ 
  type: 'FeatureCollection', 
  features: props.forteAtms.map(a => ({ 
    type: 'Feature', 
    properties: a, 
    geometry: { type: 'Point', coordinates: [Number(a.lng), Number(a.lat)] } 
  })) 
}))

const competitorsFC = computed(() => ({ 
  type: 'FeatureCollection', 
  features: props.competitorAtms.map(a => ({ 
    type: 'Feature', 
    properties: a, 
    geometry: { type: 'Point', coordinates: [Number(a.lng), Number(a.lat)] } 
  })) 
}))

// ДОБАВЛЕНА ГЕОМЕТРИЯ ИИ
const recsFC = computed(() => ({
  type: 'FeatureCollection',
  features: (props.recommendations || []).map(r => ({ 
    type: 'Feature', 
    properties: r, 
    geometry: { type: 'Point', coordinates: [Number(r.lng), Number(r.lat)] } 
  }))
}))

function updateSources() {
  if (!map || !map.isStyleLoaded()) return
  
  const sForte = map.getSource('forte') as maplibregl.GeoJSONSource
  if (sForte) sForte.setData(forteFC.value as any)

  const sComp = map.getSource('competitors') as maplibregl.GeoJSONSource
  if (sComp) sComp.setData(competitorsFC.value as any)

  // ОБНОВЛЕНИЕ ИСТОЧНИКА ИИ
  const sRecs = map.getSource('recommendations') as maplibregl.GeoJSONSource
  if (sRecs) sRecs.setData(recsFC.value as any)

  if (props.heatmapGrid) {
    const sGrid = map.getSource('grid') as maplibregl.GeoJSONSource
    if (sGrid) sGrid.setData(props.heatmapGrid)
  }

  if (props.trafficGeoJson) {
    const sTraffic = map.getSource('traffic') as maplibregl.GeoJSONSource
    if (sTraffic) sTraffic.setData(props.trafficGeoJson)
  }
}

function applyVisibility() {
  if (!map) return
  if (map.getLayer('grid-fill')) map.setLayoutProperty('grid-fill', 'visibility', props.showHeatmap ? 'visible' : 'none')
  if (map.getLayer('grid-outline')) map.setLayoutProperty('grid-outline', 'visibility', props.showHeatmap ? 'visible' : 'none')
  if (map.getLayer('traffic-lines')) map.setLayoutProperty('traffic-lines', 'visibility', props.showTraffic ? 'visible' : 'none')
  
  // УПРАВЛЕНИЕ ВИДИМОСТЬЮ ИИ
  if (map.getLayer('rec-points')) map.setLayoutProperty('rec-points', 'visibility', props.showRecommendations ? 'visible' : 'none')
  if (map.getLayer('rec-glow')) map.setLayoutProperty('rec-glow', 'visibility', props.showRecommendations ? 'visible' : 'none')
}

onMounted(() => {
  const TILE_URL = 'https://basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}.png';

  map = new maplibregl.Map({
    container: mapEl.value as HTMLDivElement,
    style: {
      version: 8,
      sources: {
        'color-tiles': { type: 'raster', tiles: [TILE_URL], tileSize: 256, attribution: 'CartoDB' },
      },
      layers: [{ id: 'base-layer', type: 'raster', source: 'color-tiles', minzoom: 0, maxzoom: 22 }]
    },
    center: [71.43, 51.13],
    zoom: 11,
    minZoom: 9,
    maxZoom: 22,
    attributionControl: false
  })

  map.on('load', async () => {
    
    // --- 0. ГРАНИЦЫ (ЖИРНЫЕ И ЯРКИЕ) ---
    try {
      const response = await fetch('https://raw.githubusercontent.com/wmgeolab/geoBoundaries/main/releaseData/gbOpen/KAZ/ADM2/geoBoundaries-KAZ-ADM2.geojson')
      const data = await response.json()
      
      const astanaDistricts = {
        type: 'FeatureCollection',
        features: data.features.filter((f: any) => 
          f.properties.shapeGroup === 'Nur-Sultan' || 
          f.properties.shapeName === 'Astana' ||
          f.properties.shapeName === 'Nur-Sultan'
        )
      }

      map!.addSource('districts', { type: 'geojson', data: astanaDistricts as any })

      map!.addLayer({
        id: 'districts-fill', type: 'fill', source: 'districts',
        paint: {
          'fill-color': ['match', ['get', 'shapeName'], 'Almaty', '#3b82f6', 'Yessil', '#10b981', 'Saryarka', '#f59e0b', 'Baikonur', '#8b5cf6', '#64748b'],
          'fill-opacity': 0.15 
        }
      })
      map!.addLayer({
        id: 'districts-line', type: 'line', source: 'districts',
        paint: { 'line-color': '#4c1d95', 'line-width': 2, 'line-dasharray': [2, 2] }
      })
      map!.addLayer({
        id: 'city-border-halo', type: 'line', source: 'districts',
        paint: { 'line-color': '#ffffff', 'line-width': 6, 'line-opacity': 0.8 }
      })
      map!.addLayer({
        id: 'city-border', type: 'line', source: 'districts',
        paint: { 'line-color': '#000000', 'line-width': 3 }
      })
      
    } catch (e) {
      console.error('Ошибка границ:', e)
    }

    // --- 1. СЕТКА ---
    map!.addSource('grid', { type: 'geojson', data: props.heatmapGrid || { type: 'FeatureCollection', features: [] } })
    map!.addLayer({
      id: 'grid-fill', type: 'fill', source: 'grid',
      paint: {
        'fill-color': ['interpolate', ['linear'], ['get', 'weight'], 0, 'rgba(0,0,0,0)', 0.3, '#10b981', 0.6, '#fcd34d', 0.9, '#ef4444'],
        'fill-opacity': 0.5
      }
    })
    map!.addLayer({
      id: 'grid-outline', type: 'line', source: 'grid',
      paint: { 'line-color': 'rgba(255,255,255,0.3)', 'line-width': 1 }
    })

    // --- 2. ТРАФИК ---
    map!.addSource('traffic', { type: 'geojson', data: props.trafficGeoJson || { type: 'FeatureCollection', features: [] } })
    map!.addLayer({
      id: 'traffic-lines', type: 'line', source: 'traffic',
      layout: { 'line-join': 'round', 'line-cap': 'round' },
      paint: { 'line-width': 5, 'line-color': ['step', ['get', 'traffic'], '#4caf50', 500, '#fcd34d', 1500, '#ef4444'], 'line-opacity': 1.0 }
    })

    // --- 3. КОНКУРЕНТЫ (ЦВЕТ ИЗМЕНЕН НА #9E9E9E) ---
    map!.addSource('competitors', { type: 'geojson', data: competitorsFC.value as any })
    map!.addLayer({
      id: 'competitor-points', type: 'circle', source: 'competitors',
      paint: { 'circle-radius': 6, 'circle-color': '#9E9E9E', 'circle-stroke-width': 2, 'circle-stroke-color': '#fff' }
    })

    // --- ДОБАВЛЕНЫ РЕКОМЕНДАЦИИ ИИ ---
    map!.addSource('recommendations', { type: 'geojson', data: recsFC.value as any })
    map!.addLayer({
      id: 'rec-glow', type: 'circle', source: 'recommendations',
      paint: { 'circle-radius': 18, 'circle-color': '#c084fc', 'circle-opacity': 0.4 }
    })
    map!.addLayer({
      id: 'rec-points', type: 'circle', source: 'recommendations',
      paint: { 'circle-radius': 8, 'circle-color': '#a855f7', 'circle-stroke-width': 2, 'circle-stroke-color': '#fff' }
    })

    // --- 4. FORTE ---
    map!.addSource('forte', { type: 'geojson', data: forteFC.value as any })
    map!.addLayer({
      id: 'forte-points', type: 'circle', source: 'forte',
      paint: {
        'circle-radius': 11,
        'circle-color': ['match', ['get', 'efficiencyStatus'], 'Effective', '#10b981', 'Ineffective', '#ef4444', '#00B8D9'],
        'circle-stroke-width': 3, 'circle-stroke-color': '#fff', 'circle-stroke-opacity': 1
      }
    })

    // СОБЫТИЯ
    map!.on('click', (e) => {
      const f = map!.queryRenderedFeatures(e.point, { layers: ['forte-points'] });
      if (f.length) { emit('select-atm', { ...f[0].properties }); return; }
      const c = map!.queryRenderedFeatures(e.point, { layers: ['competitor-points'] });
      if (c.length) { emit('select-atm', c[0].properties); return; }
      emit('select-atm', null);
    })

    map!.on('mousemove', (e) => {
      // ПРОВЕРКА ХОВЕРА ДЛЯ ИИ (Добавлена первой)
      if (props.showRecommendations) {
        const recs = map!.queryRenderedFeatures(e.point, { layers: ['rec-points', 'rec-glow'] });
        if (recs.length) {
          hoverInfo.value = { x: e.point.x + 20, y: e.point.y, type: 'recommendation', props: recs[0].properties };
          map!.getCanvas().style.cursor = 'pointer'; return;
        }
      }

      // ATM
      const atms = map!.queryRenderedFeatures(e.point, { layers: ['forte-points', 'competitor-points'] });
      if (atms.length) {
        hoverInfo.value = { x: e.point.x + 20, y: e.point.y, type: 'atm', props: atms[0].properties };
        map!.getCanvas().style.cursor = 'pointer'; return;
      }
      
      // Traffic
      if (props.showTraffic) {
        const t = map!.queryRenderedFeatures(e.point, { layers: ['traffic-lines'] });
        if (t.length) { 
          hoverInfo.value = { x: e.point.x + 20, y: e.point.y, type: 'traffic', props: { traffic: t[0].properties.traffic } }; 
          map!.getCanvas().style.cursor = 'help'; return; 
        }
      }

      // Grid
      if (props.showHeatmap) {
         const g = map!.queryRenderedFeatures(e.point, { layers: ['grid-fill'] });
         if (g.length) {
           const w = g[0].properties.weight;
           const c = w > 0.6 ? '#ef4444' : w > 0.3 ? '#fcd34d' : '#10b981';
           hoverInfo.value = { x: e.point.x + 20, y: e.point.y, type: 'grid', weight: w, color: c };
           map!.getCanvas().style.cursor = 'crosshair'; return;
         }
      }

      // Districts
      const d = map!.queryRenderedFeatures(e.point, { layers: ['districts-fill'] });
      if (d.length) {
         hoverInfo.value = { x: e.point.x + 20, y: e.point.y, type: 'district', props: d[0].properties };
         return; 
      }

      hoverInfo.value = null; map!.getCanvas().style.cursor = '';
    })

    map!.on('mouseleave', () => { hoverInfo.value = null })
    
    updateSources()
    applyVisibility()
  })
})

// ДОБАВЛЕН props.recommendations В WATCH
watch([() => props.forteAtms, () => props.competitorAtms, () => props.trafficGeoJson, () => props.heatmapGrid, () => props.recommendations], updateSources, { deep: true })
watch([() => props.showTraffic, () => props.showHeatmap, () => props.showRecommendations], applyVisibility)

onBeforeUnmount(() => { map?.remove(); map = null })
</script>

<style scoped>
.map { width: 100%; height: 100%; border-radius: 12px; overflow: hidden; background: #e2e8f0; }

.tooltip { 
  position: absolute; 
  background: rgba(15, 23, 42, 0.95); 
  backdrop-filter: blur(8px);
  border: 1px solid #334155; 
  color: #cbd5e1; 
  padding: 12px; 
  border-radius: 8px; 
  pointer-events: none; 
  z-index: 100; 
  font-size: 13px; 
  width: 230px; 
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
}
.tooltip .name { font-weight: 800; color: #fff; margin-bottom: 8px; font-size: 14px; border-bottom: 1px solid #334155; padding-bottom: 6px; }
.row { display: flex; justify-content: space-between; margin-bottom: 4px; }
.sep { height: 1px; background: #334155; margin: 8px 0; }
</style>