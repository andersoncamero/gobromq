#!/bin/bash

# 🧪 GoBroMQ - Test de Sensores IoT (Benchmark)
# =============================================

BROKER_HOST="localhost"
BROKER_PORT="1884"
USERNAME="admin"
PASSWORD="password123"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}🏠 Test de Sensores IoT - GoBroMQ${NC}"
echo "=================================="

# Configuración de sensores
declare -a SENSORS=(
    "puerta_principal:sensors/door/main"
    "puerta_trasera:sensors/door/back"
    "ventana_sala:sensors/window/living"
    "ventana_cocina:sensors/window/kitchen"
    "temperatura_sala:sensors/temp/living"
    "temperatura_cocina:sensors/temp/kitchen"
    "humedad_sala:sensors/humidity/living"
    "movimiento_entrada:sensors/motion/entrance"
    "movimiento_pasillo:sensors/motion/hallway"
    "luz_jardin:sensors/light/garden"
    "agua_tanque:sensors/water/tank"
    "gas_cocina:sensors/gas/kitchen"
)

# Función para generar datos de sensor
generate_sensor_data() {
    local sensor_name=$1
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    local value
    local status
    local unit=""
    
    case $sensor_name in
        *door*|*window*)
            value=$((RANDOM % 2))
            status=$([ $value -eq 1 ] && echo "open" || echo "closed")
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"binary\"}"
            ;;
        *temp*)
            value=$((18 + RANDOM % 15))  # 18-32°C
            status="normal"
            [ $value -lt 20 ] && status="cold"
            [ $value -gt 28 ] && status="hot"
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"unit\":\"°C\",\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"temperature\"}"
            ;;
        *humidity*)
            value=$((30 + RANDOM % 50))  # 30-80%
            status="normal"
            [ $value -lt 40 ] && status="low"
            [ $value -gt 70 ] && status="high"
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"unit\":\"%\",\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"humidity\"}"
            ;;
        *motion*)
            value=$((RANDOM % 2))
            status=$([ $value -eq 1 ] && echo "detected" || echo "clear")
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"motion\"}"
            ;;
        *light*)
            value=$((RANDOM % 1000))  # 0-999 lux
            status="normal"
            [ $value -lt 100 ] && status="dark"
            [ $value -gt 700 ] && status="bright"
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"unit\":\"lux\",\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"light\"}"
            ;;
        *water*)
            value=$((20 + RANDOM % 80))  # 20-100%
            status="normal"
            [ $value -lt 30 ] && status="low"
            [ $value -gt 90 ] && status="full"
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"unit\":\"%\",\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"level\"}"
            ;;
        *gas*)
            value=$((RANDOM % 50))  # 0-49 ppm
            status="safe"
            [ $value -gt 30 ] && status="warning"
            [ $value -gt 40 ] && status="danger"
            echo "{\"sensor\":\"$sensor_name\",\"value\":$value,\"unit\":\"ppm\",\"status\":\"$status\",\"timestamp\":\"$timestamp\",\"type\":\"gas\"}"
            ;;
    esac
}

# Test de velocidad de publicación con sensores IoT
echo -e "\n${BLUE}📊 Test de Throughput - Sensores IoT${NC}"
echo "======================================"

MESSAGE_COUNT=60  # 5 rondas de 12 sensores cada una
TOTAL_SENSORS=${#SENSORS[@]}

echo -e "${YELLOW}Enviando datos de $TOTAL_SENSORS sensores...${NC}"
echo -e "${CYAN}Sensores disponibles: ${TOTAL_SENSORS}${NC}"

START_TIME=$(date +%s.%N)

success_count=0
round=1

for i in $(seq 1 $MESSAGE_COUNT); do
    # Seleccionar sensor de forma cíclica
    sensor_index=$(( (i-1) % TOTAL_SENSORS ))
    sensor_info="${SENSORS[$sensor_index]}"
    sensor_name=$(echo "$sensor_info" | cut -d: -f1)
    sensor_topic=$(echo "$sensor_info" | cut -d: -f2)
    
    # Generar datos del sensor
    sensor_data=$(generate_sensor_data "$sensor_name")
    
    if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
        -u $USERNAME -P $PASSWORD \
        -t "$sensor_topic" -m "$sensor_data" \
        -q 0 -i "sensor_${sensor_name}_${i}_$$" 2>/dev/null; then
        ((success_count++))
    fi
    
    # Mostrar progreso cada ronda completa de sensores
    if [ $((i % TOTAL_SENSORS)) -eq 0 ]; then
        echo -e "${GREEN}🏠 Ronda $round completada ($i mensajes enviados)${NC}"
        ((round++))
    fi
    
    # Pequeña pausa para simular frecuencia real de sensores
    sleep 0.1
done

END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)
RATE=$(echo "scale=2; $success_count / $DURATION" | bc)

echo -e "\n${BLUE}📈 Resultados del Benchmark IoT:${NC}"
echo "================================="
echo -e "${GREEN}✅ Mensajes de sensores exitosos: $success_count/$MESSAGE_COUNT${NC}"
echo -e "${GREEN}🏠 Sensores simulados: $TOTAL_SENSORS tipos diferentes${NC}"
echo -e "${GREEN}⏱️  Tiempo total: ${DURATION}s${NC}"
echo -e "${GREEN}🚀 Velocidad: ${RATE} mensajes/segundo${NC}"
echo -e "${CYAN}📡 Promedio por sensor: $(echo "scale=2; $success_count / $TOTAL_SENSORS" | bc) mensajes/sensor${NC}"

# Test de latencia con sensores críticos
echo -e "\n${BLUE}⚡ Test de Latencia - Sensores Críticos${NC}"
echo "======================================="

CRITICAL_SENSORS=("puerta_principal" "movimiento_entrada" "gas_cocina")
LATENCY_TESTS=5
total_time=0

echo -e "${YELLOW}Probando latencia en sensores críticos...${NC}"

for i in $(seq 1 $LATENCY_TESTS); do
    # Seleccionar sensor crítico aleatorio
    sensor_index=$((RANDOM % ${#CRITICAL_SENSORS[@]}))
    sensor_name="${CRITICAL_SENSORS[$sensor_index]}"
    
    # Generar dato de emergencia
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    emergency_data="{\"sensor\":\"$sensor_name\",\"value\":1,\"status\":\"EMERGENCY\",\"timestamp\":\"$timestamp\",\"type\":\"alert\",\"priority\":\"high\"}"
    
    start=$(date +%s.%N)
    mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
        -u $USERNAME -P $PASSWORD \
        -t "alerts/${sensor_name}" -m "$emergency_data" \
        -q 1 -i "emergency_${sensor_name}_${i}_$$" 2>/dev/null
    end=$(date +%s.%N)
    
    latency=$(echo "$end - $start" | bc)
    total_time=$(echo "$total_time + $latency" | bc)
    
    echo -e "${GREEN}� $sensor_name: ${latency}s${NC}"
done

avg_latency=$(echo "scale=4; $total_time / $LATENCY_TESTS" | bc)
echo -e "${BLUE}📊 Latencia promedio de alertas: ${avg_latency}s${NC}"

# Mostrar ejemplo de datos generados
echo -e "\n${PURPLE}📋 Ejemplo de datos de sensores generados:${NC}"
echo "============================================="
for i in {0..4}; do
    sensor_info="${SENSORS[$i]}"
    sensor_name=$(echo "$sensor_info" | cut -d: -f1)
    sensor_topic=$(echo "$sensor_info" | cut -d: -f2)
    example_data=$(generate_sensor_data "$sensor_name")
    echo -e "${CYAN}📡 $sensor_topic:${NC}"
    echo -e "${YELLOW}   $example_data${NC}"
done

echo -e "\n${GREEN}🎉 Test de sensores IoT completado${NC}"
echo -e "${CYAN}💡 Los datos están en formato JSON con timestamp, valor y estado${NC}"
