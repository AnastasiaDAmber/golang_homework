#!/bin/bash
# Скрипт для тестирования функциональности задания 14

set -e

echo "=== Тестирование задания 14 ==="
echo ""

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Функция для проверки результата
check_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ $1${NC}"
    else
        echo -e "${RED}✗ $1${NC}"
        exit 1
    fi
}

echo "1. Запуск unit-тестов для rabbitmq..."
go test ./internal/rabbitmq -v
check_result "Тесты rabbitmq"

echo ""
echo "2. Запуск unit-тестов для storage (memory) - новые методы..."
go test ./internal/storage/memory -v -run "TestStorage_(EventsToNotify|DeleteOldEvents)"
check_result "Тесты storage/memory"

echo ""
echo "3. Запуск unit-тестов для app - новые методы..."
go test ./internal/app -v -run "TestApp_(EventsToNotify|DeleteOldEvents)"
check_result "Тесты app"

echo ""
echo "4. Запуск всех тестов с флагом -race..."
go test -race ./internal/...
check_result "Все тесты с race detector"

echo ""
echo "5. Проверка сборки всех сервисов..."
make build
check_result "Сборка всех сервисов"

echo ""
echo "6. Проверка линтера..."
make lint
check_result "Линтер"

echo ""
echo -e "${GREEN}=== Все тесты пройдены успешно! ===${NC}"

