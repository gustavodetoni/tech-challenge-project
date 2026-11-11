#!/usr/bin/env bash
set -euo pipefail

OS_URL="${OS_URL:-http://localhost:8081}"
BILLING_URL="${BILLING_URL:-http://localhost:8082}"
EXECUTION_URL="${EXECUTION_URL:-http://localhost:8083}"

command -v jq >/dev/null || { echo "jq is required"; exit 1; }

CLIENT_ID="${CLIENT_ID:-$(uuidgen)}"
VEHICLE_ID="${VEHICLE_ID:-$(uuidgen)}"

echo "1. Criando OS"
ORDER_JSON="$(curl -fsS -X POST "$OS_URL/service-orders" -H 'Content-Type: application/json' -d "{\"client_id\":\"$CLIENT_ID\",\"vehicle_id\":\"$VEHICLE_ID\",\"amount_cents\":15000}")"
echo "$ORDER_JSON" | jq .
ORDER_ID="$(echo "$ORDER_JSON" | jq -r .id)"

sleep 1
echo "2. Criando orçamento"
QUOTE_JSON="$(curl -fsS -X POST "$BILLING_URL/quotes" -H 'Content-Type: application/json' -d "{\"order_id\":\"$ORDER_ID\",\"amount_cents\":15000}")"
echo "$QUOTE_JSON" | jq .
QUOTE_ID="$(echo "$QUOTE_JSON" | jq -r .id)"

echo "3. Criando pagamento"
PAYMENT_JSON="$(curl -fsS -X POST "$BILLING_URL/payments" -H 'Content-Type: application/json' -d "{\"quote_id\":\"$QUOTE_ID\"}")"
echo "$PAYMENT_JSON" | jq .
PAYMENT_ID="$(echo "$PAYMENT_JSON" | jq -r .payment_id)"

echo "4. Simulando aprovação do Mercado Pago"
curl -fsS -X POST "$BILLING_URL/webhooks/mercadopago" -H 'Content-Type: application/json' -d "{\"type\":\"payment.approved\",\"status\":\"approved\",\"data\":{\"id\":\"$PAYMENT_ID\"}}" | jq .

sleep 2
echo "5. Consultando execução liberada pela mensageria"
curl -fsS "$EXECUTION_URL/executions/$ORDER_ID" | jq .

echo "6. Finalizando execução"
curl -fsS -X POST "$EXECUTION_URL/executions/$ORDER_ID/finish" | jq .
sleep 1
curl -fsS "$OS_URL/service-orders/$ORDER_ID" | jq .

echo "7. Demonstração de compensação"
echo "Para demonstrar a falha automática, crie uma segunda OS, aguarde o pagamento e execute antes da finalização:"
echo "curl -X POST $EXECUTION_URL/executions/<ORDER_ID>/fail"
echo "O Execution Service publicará ExecutionFailed; o OS Service cancelará a OS e o Billing Service solicitará o estorno."
