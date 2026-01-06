# Развертывание Calendar в Kubernetes

## Быстрый старт

### 1. Установка Kubernetes кластера

**Minikube:**
```bash
minikube start
eval $(minikube docker-env)
```

**k3s:**
```bash
curl -sfL https://get.k3s.io | sh -
```

**MicroK8s:**
```bash
sudo snap install microk8s --classic
microk8s enable dns storage ingress
```

### 2. Подготовка образов (для minikube)

```bash
docker build -f Dockerfile.calendar -t calendar:latest .
docker build -f Dockerfile.scheduler -t calendar-scheduler:latest .
docker build -f Dockerfile.sender -t calendar-sender:latest .
docker build -f Dockerfile.migrate -t calendar-migrate:latest .
```

### 3. Развертывание с Helm

```bash
# Установка
helm install calendar .

# Проверка
kubectl get pods
kubectl get services

# Доступ через port-forward
kubectl port-forward svc/calendar-calendar 8888:8888
```

### 4. Проверка работы

```bash
curl http://localhost:8888/
curl http://localhost:8888/api/events
```

## Структура Helm Chart

- `Chart.yaml` - метаданные chart
- `values.yaml` - дефолтные значения
- `templates/` - Kubernetes манифесты:
  - `deployment-calendar.yaml` - API сервер
  - `deployment-scheduler.yaml` - Планировщик
  - `deployment-sender.yaml` - Отправитель уведомлений
  - `deployment-rabbitmq.yaml` - RabbitMQ
  - `statefulset-postgresql.yaml` - PostgreSQL
  - `service-*.yaml` - Сервисы для всех компонентов
  - `ingress.yaml` - Ingress для внешнего доступа
  - `job-migrate.yaml` - Job для миграций БД
  - `_helpers.tpl` - Вспомогательные шаблоны

Подробная документация: `docs/KUBERNETES_DEPLOYMENT.md`

