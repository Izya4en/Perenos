# 1. Используем Node 22 (Alpine)
FROM node:22-alpine

# 2. Устанавливаем рабочую папку
WORKDIR /app

# 3. Устанавливаем инструменты для сборки (чтобы не было ошибок node-gyp)
RUN apk add --no-cache python3 make g++

# 4. Копируем файлы зависимостей
COPY package*.json ./

# 5. Устанавливаем пакеты
RUN npm install

# 6. Копируем весь проект
COPY . .

# 7. Собираем проект
RUN npm run build

# 8. Настройки порта и хоста
EXPOSE 3000
ENV NUXT_HOST=0.0.0.0
ENV NUXT_PORT=3000

# 9. Запуск
CMD ["node", ".output/server/index.mjs"]