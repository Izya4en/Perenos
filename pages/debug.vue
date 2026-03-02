<template>
  <div style="padding: 20px; color: white; background: #000; min-height: 100vh;">
    <h1>РЕЖИМ ОТЛАДКИ</h1>
    
    <div style="margin-bottom: 20px;">
      <button @click="refresh" style="padding: 10px; background: cyan;">Обновить данные</button>
    </div>

    <div v-if="error" style="border: 2px solid red; padding: 20px; color: red;">
      <h3>ОШИБКА ЗАГРУЗКИ:</h3>
      <pre>{{ error }}</pre>
    </div>

    <div v-else-if="data">
      <h3>Успешно получено {{ Array.isArray(data) ? data.length : '?' }} объектов</h3>
      
      <p>Вот как выглядит ПЕРВЫЙ элемент (сырой JSON):</p>
      <pre style="background: #333; padding: 10px; overflow: auto;">{{ JSON.stringify(data[0], null, 2) }}</pre>

      <p>Вот полный ответ сервера:</p>
      <textarea style="width: 100%; height: 200px; color: black;">{{ JSON.stringify(data, null, 2) }}</textarea>
    </div>

    <div v-else>
      Загрузка...
    </div>
  </div>
</template>

<script setup>
// Пробуем прямой запрос, минуя прокси, если сервер разрешает CORS,
// Если нет - пробуем через прокси
const { data, error, refresh } = await useFetch('/api/traffic', { 
  server: false 
})
</script>