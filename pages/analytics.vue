<template>
  <div class="page-analytics">
    <h2>Сводная Аналитика</h2>
    
    <div v-if="error" class="loading red">
      ❌ Ошибка связи с сервером: {{ error.message }}
    </div>

    <div v-else-if="pending" class="loading">
      ⏳ Загрузка данных...
    </div>

    <div v-else-if="apiData" class="content-wrapper">
      <div class="grid-stats">
        <div class="card">
          <h3>Количество (ATM)</h3>
          <div class="chart-bars">
            <div v-for="b in bankStats" :key="b.name" class="bar-row">
              <span class="b-label">{{ b.name }}</span>
              <div class="b-track">
                <div class="b-fill" :style="{ width: b.pct + '%', background: b.color }"></div>
              </div>
              <span class="b-val">{{ b.count }}</span>
            </div>
          </div>
        </div>

        <div class="card">
          <h3>Статистика Сети Forte</h3>
          <div class="kpi-grid">
            <div class="kpi-box">
              <div class="k-val green">{{ efficiencyStats.effective }}</div>
              <div class="k-lbl">Эффективные</div>
            </div>
            <div class="kpi-box">
              <div class="k-val red">{{ efficiencyStats.ineffective }}</div>
              <div class="k-lbl">Проблемные</div>
            </div>
          </div>
        </div>
      </div>

      <div class="card table-card">
        <h3>Детальный список АТМ</h3>
        <div class="table-scroll-area">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Район</th>
                <th>Адрес</th>
                <th>Баланс</th>
                <th>Статус</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="atm in forte" :key="atm.id || atm.terminal_id">
                <td class="id-col clickable-id" @click="goToMap(atm)">
                  {{ atm.terminal_id || atm.name }}
                </td>
                
                <td><span class="district-tag">{{ determineDistrict(atm) }}</span></td>
                <td>{{ atm.address || 'Не указан' }}</td>
                <td class="mono">{{ fmtKZT(atm.totalCashKZT || 0) }}</td>
                <td>
                  <span class="badge" :class="getBadgeClass(atm)">
                    {{ translateStatus(atm) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

// БЕЗОПАСНЫЙ ЗАПРОС (добавили pending и error)
const { data: apiData, pending, error } = await useFetch<any>('http://localhost:8080/api/dashboard', {
  server: false, // Важно для клиентского рендеринга
  key: 'analytics-data'
})

const forte = computed(() => apiData.value?.forte || [])
const competitors = computed(() => apiData.value?.competitors || [])

const bankStats = computed(() => {
  const counts: Record<string, number> = { 'Forte': forte.value.length }
  competitors.value.forEach((c: any) => counts[c.bank] = (counts[c.bank] || 0) + 1)
  const total = forte.value.length + competitors.value.length
  
  if (total === 0) return [] // Защита от деления на ноль

  return Object.entries(counts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5)
    .map(([name, count]) => ({ 
      name, 
      count, 
      pct: (count / total) * 100, 
      color: name === 'Forte' ? '#00B8D9' : '#64748b' 
    }))
})

const efficiencyStats = computed(() => {
  let e = 0, i = 0
  forte.value.forEach((a: any) => { 
    if (a.efficiencyStatus === 'Effective') e++ 
    else if (a.efficiencyStatus === 'Ineffective') i++ 
  })
  return { effective: e, ineffective: i }
})

function goToMap(atm: any) {
  const id = atm.terminal_id || atm.id
  if (id) {
    router.push({ path: '/', query: { atmId: id } })
  }
}

function fmtKZT(n: number) { 
  return new Intl.NumberFormat('ru-KZ', { style: 'currency', currency: 'KZT', maximumFractionDigits: 0 }).format(n || 0) 
}

function determineDistrict(atm: any) {
  const addr = (atm.address || '').toLowerCase()
  if (addr.includes('достык') || addr.includes('expo') || addr.includes('mega') || addr.includes('abu')) return 'Есиль'
  if (addr.includes('хан шатыр') || addr.includes('keruen') || addr.includes('туран')) return 'Нура'
  if (addr.includes('республики') || addr.includes('вокзал') || addr.includes('алаш')) return 'Сарыарка'
  if (addr.includes('алматы') || addr.includes('момышулы') || addr.includes('тауелсиздик')) return 'Алматы'
  if (addr.includes('молдагуловой') || addr.includes('потанина')) return 'Байконур'
  return 'Есиль'
}

function getBadgeClass(atm: any) { 
  return (atm.efficiencyStatus || '').toLowerCase() === 'effective' ? 'Effective' : 'Ineffective' 
}

function translateStatus(atm: any) { 
  return (atm.efficiencyStatus || '').toLowerCase() === 'effective' ? 'Активен' : 'Проблема' 
}
</script>

<style scoped>
.page-analytics { padding: 20px; max-width: 1200px; margin: 0 auto; height: calc(100vh - 60px); overflow-y: auto; }
h2 { color: #fff; margin-bottom: 20px; border-bottom: 1px solid #334155; padding-bottom: 10px; }
.content-wrapper { display: flex; flex-direction: column; gap: 20px; }
.grid-stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
.card { background: #0f172a; border: 1px solid #1e293b; border-radius: 12px; padding: 20px; }
.card h3 { margin: 0 0 15px 0; color: #94a3b8; font-size: 14px; text-transform: uppercase; }
.bar-row { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
.b-label { width: 80px; font-size: 12px; color: #cbd5e1; }
.b-track { flex: 1; height: 8px; background: #1e293b; border-radius: 4px; overflow: hidden; }
.b-fill { height: 100%; border-radius: 4px; }
.b-val { font-size: 12px; color: #fff; width: 30px; text-align: right; }
.kpi-grid { display: flex; justify-content: space-around; text-align: center; }
.k-val { font-size: 32px; font-weight: bold; }
.k-lbl { font-size: 12px; color: #64748b; }
.green { color: #10b981; } .red { color: #ef4444; }
.table-card { padding: 0; overflow: hidden; display: flex; flex-direction: column; }
.table-card h3 { margin: 20px 20px 10px 20px; }
.table-scroll-area { max-height: 400px; overflow-y: auto; padding: 0 20px 20px 20px; }
.table-scroll-area::-webkit-scrollbar { width: 6px; }
.table-scroll-area::-webkit-scrollbar-track { background: #0f172a; }
.table-scroll-area::-webkit-scrollbar-thumb { background: #334155; border-radius: 3px; }
table { width: 100%; color: #e2e8f0; border-collapse: collapse; font-size: 13px; }
thead th { text-align: left; padding: 12px 10px; color: #94a3b8; border-bottom: 1px solid #334155; position: sticky; top: 0; background: #0f172a; z-index: 10; }
td { padding: 12px 10px; border-bottom: 1px solid #1e293b; }
tr:last-child td { border-bottom: none; }
.id-col { color: #00B8D9; font-weight: bold; }
.district-tag { background: rgba(148, 163, 184, 0.1); color: #cbd5e1; padding: 2px 6px; border-radius: 4px; font-size: 11px; }
.mono { font-family: monospace; }
.badge { padding: 4px 10px; border-radius: 20px; font-size: 11px; font-weight: bold; }
.badge.Effective { color: #10b981; background: rgba(16,185,129,0.15); }
.badge.Ineffective { color: #ef4444; background: rgba(239,68,68,0.15); }
.loading { color: #fff; text-align: center; margin-top: 50px; font-size: 18px; }
.loading.red { color: #ef4444; }
.clickable-id { cursor: pointer; text-decoration: underline; transition: color 0.2s; }
.clickable-id:hover { color: #fff; }
</style>