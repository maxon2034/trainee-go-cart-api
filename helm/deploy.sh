#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHART_DIR="${SCRIPT_DIR}/cart-api"
RELEASE="cart-api"
APP_IMAGE="trainee-go-cart-api-app:latest"
DB_IMAGE="postgres:16-alpine"
LOCAL_PORT="${LOCAL_PORT:-8080}"

step() {
  echo
  echo "=================================================="
  echo ">>> $1"
  echo "=================================================="
}

on_error() {
  echo
  echo "!!! Ошибка на строке $1. Диагностика:"
  if [[ -n "${NAMESPACE:-}" ]]; then
    kubectl get pods -n "${NAMESPACE}" 2>/dev/null || true
    echo "Подробности: kubectl describe pod -l app=${RELEASE} -n ${NAMESPACE}"
    echo "Логи:        kubectl logs deploy/${RELEASE} -n ${NAMESPACE}"
  fi
}
trap 'on_error $LINENO' ERR

PASSWORD_FILE=""
cleanup() {
  if [[ -n "${PASSWORD_FILE}" && -f "${PASSWORD_FILE}" ]]; then
    rm -f "${PASSWORD_FILE}"
  fi
}
trap cleanup EXIT

step "Проверка необходимых утилит"
for tool in docker minikube kubectl helm; do
  if ! command -v "${tool}" >/dev/null 2>&1; then
    echo "Не найдена утилита: ${tool}"
    exit 1
  fi
  echo "OK: ${tool}"
done

if [[ ! -d "${CHART_DIR}" ]]; then
  echo "Не найдена папка чарта: ${CHART_DIR}"
  exit 1
fi

step "Выбор окружения"
read -r -p "Окружение (dev/qa/prod): " ENVIRONMENT
case "${ENVIRONMENT}" in
  dev|qa|prod) ;;
  *)
    echo "Недопустимое окружение: '${ENVIRONMENT}'. Допустимо: dev, qa, prod"
    exit 1
    ;;
esac
NAMESPACE="${ENVIRONMENT}"
VALUES_FILE="${CHART_DIR}/values-${ENVIRONMENT}.yaml"
if [[ ! -f "${VALUES_FILE}" ]]; then
  echo "Не найден файл значений: ${VALUES_FILE}"
  exit 1
fi
echo "Окружение: ${ENVIRONMENT}, namespace: ${NAMESPACE}"

step "Ввод пароля от БД"
read -r -s -p "Пароль БД: " DB_PASSWORD
echo
read -r -s -p "Повторите пароль: " DB_PASSWORD_CONFIRM
echo
if [[ -z "${DB_PASSWORD}" ]]; then
  echo "Пароль не может быть пустым"
  exit 1
fi
if [[ "${DB_PASSWORD}" != "${DB_PASSWORD_CONFIRM}" ]]; then
  echo "Пароли не совпадают"
  exit 1
fi
PASSWORD_FILE="$(mktemp)"
chmod 600 "${PASSWORD_FILE}"
printf '%s' "${DB_PASSWORD}" > "${PASSWORD_FILE}"
unset DB_PASSWORD DB_PASSWORD_CONFIRM
echo "Пароль принят"

step "Проверка локальных образов на хосте"
if ! docker image inspect "${APP_IMAGE}" >/dev/null 2>&1; then
  echo "Образ ${APP_IMAGE} не найден на хосте. Сначала соберите его:"
  echo "  docker build -t ${APP_IMAGE} <папка с Dockerfile приложения>"
  exit 1
fi
echo "OK: ${APP_IMAGE}"
if ! docker image inspect "${DB_IMAGE}" >/dev/null 2>&1; then
  echo "Образ ${DB_IMAGE} не найден локально, скачиваю"
  docker pull "${DB_IMAGE}"
fi
echo "OK: ${DB_IMAGE}"

step "Запуск minikube"
minikube start

step "Включение metrics-server (нужен для HPA)"
minikube addons enable metrics-server

step "Загрузка образа приложения в minikube: ${APP_IMAGE}"
minikube image load "${APP_IMAGE}"

step "Загрузка образа БД в minikube: ${DB_IMAGE}"
minikube image load "${DB_IMAGE}"

step "Проверка, что образы видны внутри minikube"
minikube image ls | grep -E "trainee-go-cart-api-app|postgres" || {
  echo "Образы не найдены внутри minikube"
  exit 1
}

step "Проверка чарта (helm lint)"
helm lint "${CHART_DIR}" -f "${VALUES_FILE}" --set-file "secret.dbPassword=${PASSWORD_FILE}"

step "Установка/обновление релиза ${RELEASE} в namespace ${NAMESPACE}"
helm upgrade --install "${RELEASE}" "${CHART_DIR}" \
  -n "${NAMESPACE}" --create-namespace \
  -f "${VALUES_FILE}" \
  --set-file "secret.dbPassword=${PASSWORD_FILE}" \
  --wait --timeout 5m

step "Статус релиза"
helm list -n "${NAMESPACE}"
helm history "${RELEASE}" -n "${NAMESPACE}"

step "Состояние ресурсов в namespace ${NAMESPACE}"
kubectl get pods,svc,hpa,pvc -n "${NAMESPACE}"

step "Проброс порта: http://localhost:${LOCAL_PORT} -> svc/${RELEASE}:80 (Ctrl+C для остановки)"
kubectl port-forward "svc/${RELEASE}" "${LOCAL_PORT}:80" -n "${NAMESPACE}"