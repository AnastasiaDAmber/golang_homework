# Развертывание Calendar в Kubernetes

## Требования

- Kubernetes кластер (minikube, k3s, или microk8s)
- Helm 3.x
- kubectl настроен для работы с кластером

## Установка кластера Kubernetes

### Minikube

```bash
# Установка minikube (пример для Linux)
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Запуск кластера
minikube start

# Проверка
kubectl get nodes
```

### k3s

```bash
# Установка k3s
curl -sfL https://get.k3s.io | sh -

# Проверка
kubectl get nodes
```

### MicroK8s

```bash
# Установка microk8s
sudo snap install microk8s --classic

# Добавление пользователя в группу
sudo usermod -a -G microk8s $USER
newgrp microk8s

# Включение необходимых аддонов
microk8s enable dns storage ingress

# Проверка
microk8s kubectl get nodes
```

## Подготовка образов


### Для minikube

```bash
# Использовать Docker daemon из minikube
eval $(minikube docker-env)

# Собрать образы
docker build -f Dockerfile.calendar -t calendar:latest .
docker build -f Dockerfile.scheduler -t calendar-scheduler:latest .
docker build -f Dockerfile.sender -t calendar-sender:latest .
docker build -f Dockerfile.migrate -t calendar-migrate:latest .
```

### Для других кластеров

Загрузить образы в доступный registry и обновите `values.yaml`:

```yaml
calendar:
  image:
    repository: your-registry/calendar
    tag: "latest"
```

## Развертывание с помощью Helm

### 1. Проверка манифестов

```bash
# Проверка шаблонов
helm template calendar-chart .

# Проверка синтаксиса
helm lint .
```

### 2. Установка chart

```bash
# Установка в namespace по умолчанию
helm install calendar calendar-chart .

# Или в отдельный namespace
kubectl create namespace calendar
helm install calendar calendar-chart . -n calendar
```

### 3. Проверка развертывания

```bash
# Проверка подов
kubectl get pods

# Проверка сервисов
kubectl get services

# Проверка развертываний
kubectl get deployments

# Проверка StatefulSet
kubectl get statefulsets

# Проверка миграций
kubectl get jobs
```

### 4. Просмотр логов

```bash
# Логи calendar
kubectl logs -l app.kubernetes.io/component=calendar

# Логи scheduler
kubectl logs -l app.kubernetes.io/component=scheduler

# Логи sender
kubectl logs -l app.kubernetes.io/component=sender
```

### 5. Доступ к приложению

#### Через Port Forward

```bash
# Проброс порта для HTTP API
kubectl port-forward svc/calendar-calendar 8888:8888

# Теперь доступно на http://localhost:8888
```

#### Через Ingress

Если Ingress включен в `values.yaml`:

```yaml
ingress:
  enabled: true
  hosts:
    - host: calendar.local
      paths:
        - path: /
          pathType: Prefix
```

Добавьте запись в `/etc/hosts`:
```
<INGRESS_IP> calendar.local
```

### 6. Обновление развертывания

```bash
# После изменения values.yaml
helm upgrade calendar calendar-chart .

# Или с новыми значениями
helm upgrade calendar calendar-chart . --set calendar.replicaCount=2
```

### 7. Удаление

```bash
# Удаление release
helm uninstall calendar

# Удаление с сохранением истории
helm uninstall calendar --keep-history
```

## Структура Helm Chart

```
calendar-chart/
├── Chart.yaml              # Метаданные chart
├── values.yaml             # Дефолтные значения
├── charts/                 # Зависимости (пустая)
└── templates/              # Шаблоны манифестов
    ├── _helpers.tpl        # Вспомогательные функции
    ├── deployment-calendar.yaml
    ├── deployment-scheduler.yaml
    ├── deployment-sender.yaml
    ├── deployment-rabbitmq.yaml
    ├── statefulset-postgresql.yaml
    ├── service-calendar.yaml
    ├── service-scheduler.yaml
    ├── service-sender.yaml
    ├── service-rabbitmq.yaml
    ├── service-postgresql.yaml
    ├── ingress.yaml
    └── job-migrate.yaml
```

## Настройка values.yaml

Основные параметры для настройки:

- `calendar.replicaCount` - количество реплик API сервера
- `calendar.image.repository` и `calendar.image.tag` - образ для calendar
- `postgresql.persistence.size` - размер хранилища для БД
- `ingress.enabled` - включить/выключить Ingress
- `ingress.hosts` - хосты для Ingress

## Troubleshooting

### Поды не запускаются

```bash
# Проверить события
kubectl get events --sort-by='.lastTimestamp'

# Описание пода
kubectl describe pod <pod-name>

# Логи пода
kubectl logs <pod-name>
```

### Проблемы с подключением к БД

```bash
# Проверить сервис PostgreSQL
kubectl get svc calendar-postgresql

# Проверить поды PostgreSQL
kubectl get pods -l app.kubernetes.io/component=postgresql

# Проверить логи calendar
kubectl logs -l app.kubernetes.io/component=calendar
```

### Проблемы с RabbitMQ

```bash
# Проверить сервис RabbitMQ
kubectl get svc calendar-rabbitmq

# Проверить поды RabbitMQ
kubectl get pods -l app.kubernetes.io/component=rabbitmq

# Проверить логи scheduler и sender
kubectl logs -l app.kubernetes.io/component=scheduler
kubectl logs -l app.kubernetes.io/component=sender
```

