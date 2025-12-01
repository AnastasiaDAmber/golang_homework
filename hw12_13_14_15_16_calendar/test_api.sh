#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== Тестирование HTTP API календаря ==="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Создание события
echo "1. Создание события..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/api/events \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-1",
    "title": "Тестовое событие",
    "at": "2025-12-01T15:00:00Z",
    "duration": "1h",
    "description": "Описание тестового события",
    "user_id": "user1",
    "notify_before": "15m"
  }')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 201 ]; then
  echo -e "${GREEN}✓ Событие создано${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка создания события (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 2. Получение события
echo "2. Получение события по ID..."
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events/get?id=test-1")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ Событие получено${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка получения события (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 3. Список всех событий
echo "3. Получение списка всех событий..."
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ Список событий получен${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка получения списка (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 4. Обновление события
echo "4. Обновление события..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT $BASE_URL/api/events/update \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-1",
    "title": "Обновленное событие",
    "at": "2025-12-01T16:00:00Z",
    "duration": "2h",
    "description": "Обновленное описание",
    "user_id": "user1",
    "notify_before": "30m"
  }')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ Событие обновлено${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка обновления (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 5. События за день
echo "5. Получение событий за день..."
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events/day?day_start=2025-12-01T00:00:00Z")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ События за день получены${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка получения событий за день (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 6. События за неделю
echo "6. Получение событий за неделю..."
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events/week?week_start=2025-12-01T00:00:00Z")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ События за неделю получены${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка получения событий за неделю (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 7. События за месяц
echo "7. Получение событий за месяц..."
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/events/month?month_start=2025-12-01T00:00:00Z")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ События за месяц получены${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка получения событий за месяц (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

# 8. Удаление события
echo "8. Удаление события..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/events/delete?id=test-1")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ Событие удалено${NC}"
  echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
else
  echo -e "${RED}✗ Ошибка удаления (HTTP $HTTP_CODE)${NC}"
  echo "$BODY"
fi
echo ""

echo "=== Тестирование завершено ==="

