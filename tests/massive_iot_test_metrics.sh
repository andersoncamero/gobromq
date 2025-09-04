#!/bin/bash

# 🧪 GoBroMQ - Simulador de Tráfico Masivo IoT con Métricas Detalladas
# ===================================================================

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
WHITE='\033[1;37m'
NC='\033[0m'

echo -e "${PURPLE}🏢 Simulador de Tráfico Masivo IoT - GoBroMQ${NC}"
echo "==============================================="

# Configuración de dispositivos IoT (20 zonas para ~140 dispositivos)
declare -A ZONES=(
    ["living_room"]="sala"
    ["kitchen"]="cocina" 
    ["bedroom_1"]="dormitorio_1"
    ["bedroom_2"]="dormitorio_2"
    ["bedroom_3"]="dormitorio_3"
    ["bathroom_1"]="baño_1"
    ["bathroom_2"]="baño_2"
    ["garage"]="garaje"
    ["garden"]="jardin"
    ["entrance"]="entrada"
    ["hallway"]="pasillo"
    ["office"]="oficina"
    ["laundry"]="lavanderia"
    ["basement"]="sotano"
    ["attic"]="atico"
    ["balcony"]="balcon"
    ["storage"]="almacen"
    ["dining_room"]="comedor"
    ["guest_room"]="cuarto_invitados"
    ["gym"]="gimnasio"
)

declare -A DEVICE_TYPES=(
    ["door"]="puerta"
    ["window"]="ventana"
    ["motion"]="movimiento"
    ["temperature"]="temperatura"
    ["humidity"]="humedad"
    ["light"]="luz"
    ["switch"]="interruptor"
    ["smoke"]="humo"
    ["water"]="agua"
    ["gas"]="gas"
    ["camera"]="camara"
    ["speaker"]="altavoz"
    ["thermostat"]="termostato"
    ["blind"]="persiana"
    ["lock"]="cerradura"
    ["fan"]="ventilador"
    ["ac"]="aire_acondicionado"
    ["heater"]="calefactor"
    ["pressure"]="presion"
    ["vibration"]="vibracion"
    ["proximity"]="proximidad"
    ["sound"]="sonido"
    ["air_quality"]="calidad_aire"
    ["uv"]="uv"
    ["ph"]="ph"
)

# Función para generar datos realistas por tipo de sensor
generate_device_data() {
    local device_type=$1
    local zone=$2
    local device_id=$3
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    local battery=$((70 + RANDOM % 30))
    local signal_strength=$((-40 - RANDOM % 20))
    
    case $device_type in
        "door"|"window"|"lock")
            local state=$((RANDOM % 2))
            local status=$([ $state -eq 1 ] && echo "open" || echo "closed")
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"state\":${state},\"status\":\"${status}\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        "motion"|"camera")
            local detected=$((RANDOM % 2))
            local status=$([ $detected -eq 1 ] && echo "motion_detected" || echo "no_motion")
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"detected\":${detected},\"status\":\"${status}\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        "temperature"|"thermostat")
            local temp=$((16 + RANDOM % 20))
            local target=$((20 + RANDOM % 6))
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"temperature\":${temp},\"target\":${target},\"unit\":\"celsius\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        "humidity")
            local humidity=$((30 + RANDOM % 50))
            local status="normal"
            [ $humidity -lt 40 ] && status="low"
            [ $humidity -gt 70 ] && status="high"
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"humidity\":${humidity},\"unit\":\"percent\",\"status\":\"${status}\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        "light"|"switch")
            local brightness=$((RANDOM % 101))
            local state=$((RANDOM % 2))
            local status=$([ $state -eq 1 ] && echo "on" || echo "off")
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"state\":${state},\"brightness\":${brightness},\"status\":\"${status}\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        "smoke"|"gas")
            local level=$((RANDOM % 100))
            local status="safe"
            [ $level -gt 60 ] && status="warning"
            [ $level -gt 80 ] && status="danger"
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"level\":${level},\"unit\":\"ppm\",\"status\":\"${status}\",\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
        *)
            # Dispositivo genérico
            local value=$((RANDOM % 100))
            echo "{\"device_id\":\"${device_id}\",\"zone\":\"${zone}\",\"type\":\"${device_type}\",\"value\":${value},\"battery\":${battery},\"signal\":${signal_strength},\"timestamp\":\"${timestamp}\"}"
            ;;
    esac
}

# Función para simular dispositivo con métricas
simulate_device_with_metrics() {
    local device_id=$1
    local zone=$2
    local device_type=$3
    local messages_per_device=$4
    local delay=$5
    
    local device_start=$(date +%s.%N)
    local success_count=0
    
    for i in $(seq 1 $messages_per_device); do
        local msg_start=$(date +%s.%N)
        local data=$(generate_device_data "$device_type" "$zone" "$device_id")
        local topic="iot/${zone}/${device_type}/${device_id}"
        
        if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
            -u $USERNAME -P $PASSWORD \
            -t "$topic" -m "$data" \
            -q 0 -i "${device_id}_${i}_$$" 2>/dev/null; then
            echo "✓" > "/tmp/gobromq_success_${device_id}_${i}_$$"
            ((success_count++))
        fi
        
        local msg_end=$(date +%s.%N)
        local msg_duration=$(echo "$msg_end - $msg_start" | bc)
        echo "$device_id:$i:$msg_duration" >> "/tmp/gobromq_msg_timing_$$"
        
        sleep $delay
    done
    
    local device_end=$(date +%s.%N)
    local device_duration=$(echo "$device_end - $device_start" | bc)
    echo "$device_id:$device_duration:$success_count" >> "/tmp/gobromq_timing_$$"
}

echo -e "${CYAN}🏠 Configurando simulación de edificio inteligente...${NC}"

# Generar dispositivos estratégicamente para alcanzar ~140 dispositivos
declare -a ALL_DEVICES=()
device_count=0

for zone_key in "${!ZONES[@]}"; do
    zone_name="${ZONES[$zone_key]}"
    
    case $zone_key in
        "kitchen")
            for device_key in "temperature" "humidity" "gas" "water" "light" "motion" "smoke" "switch" "camera"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "living_room")
            for device_key in "temperature" "light" "motion" "speaker" "blind" "camera" "switch" "air_quality" "sound"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "bedroom_1"|"bedroom_2"|"bedroom_3")
            for device_key in "temperature" "light" "motion" "window" "blind" "thermostat" "switch" "air_quality"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "bathroom_1"|"bathroom_2")
            for device_key in "humidity" "temperature" "motion" "water" "light" "switch" "fan"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "entrance"|"hallway")
            for device_key in "door" "motion" "camera" "lock" "light" "switch" "proximity"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "garage")
            for device_key in "door" "motion" "camera" "light" "temperature" "vibration" "pressure" "switch"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "garden"|"balcony")
            for device_key in "temperature" "humidity" "light" "motion" "camera" "uv" "sound" "water"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        "office")
            for device_key in "temperature" "light" "motion" "air_quality" "sound" "switch" "camera"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
        *)
            # Otras zonas con dispositivos básicos
            for device_key in "temperature" "light" "motion" "switch" "humidity"; do
                device_id="${zone_key}_${device_key}_01"
                ALL_DEVICES+=("$device_id:$zone_name:$device_key")
                ((device_count++))
            done
            ;;
    esac
done

echo -e "${WHITE}📊 Dispositivos IoT configurados: ${device_count}${NC}"
echo -e "${WHITE}🏠 Zonas: ${#ZONES[@]}${NC}"
echo -e "${WHITE}📱 Tipos de dispositivos: ${#DEVICE_TYPES[@]}${NC}"

# Configuración del test
MESSAGES_PER_DEVICE=5
DEVICE_DELAY=0.2
CONCURRENT_DEVICES=10

echo -e "\n${BLUE}🚀 Iniciando simulación con métricas detalladas...${NC}"
echo "=================================================="
echo -e "${YELLOW}📈 Mensajes por dispositivo: ${MESSAGES_PER_DEVICE}${NC}"
echo -e "${YELLOW}⏱️  Delay entre mensajes: ${DEVICE_DELAY}s${NC}"
echo -e "${YELLOW}🔄 Dispositivos concurrentes: ${CONCURRENT_DEVICES}${NC}"
echo -e "${YELLOW}📊 Total estimado de mensajes: $((device_count * MESSAGES_PER_DEVICE))${NC}"

# Limpiar archivos temporales
rm -f /tmp/gobromq_success_*_$$
rm -f /tmp/gobromq_timing_$$
rm -f /tmp/gobromq_msg_timing_$$

echo -e "${YELLOW}📊 Iniciando medición de métricas...${NC}"

TOTAL_START_TIME=$(date +%s.%N)
START_TIME=$(date +%s.%N)

# Ejecutar dispositivos en lotes con métricas
declare -a BATCH_TIMES=()

for ((batch=0; batch<device_count; batch+=CONCURRENT_DEVICES)); do
    batch_number=$((batch/CONCURRENT_DEVICES + 1))
    batch_start=$(date +%s.%N)
    
    echo -e "\n${CYAN}🔄 Lote $batch_number: Dispositivos $((batch+1))-$((batch+CONCURRENT_DEVICES))${NC}"
    
    # Iniciar dispositivos en background
    for ((i=batch; i<batch+CONCURRENT_DEVICES && i<device_count; i++)); do
        device_info="${ALL_DEVICES[$i]}"
        device_id=$(echo "$device_info" | cut -d: -f1)
        zone=$(echo "$device_info" | cut -d: -f2)
        device_type=$(echo "$device_info" | cut -d: -f3)
        
        echo -e "${GREEN}🚀 ${device_id}${NC}"
        simulate_device_with_metrics "$device_id" "$zone" "$device_type" "$MESSAGES_PER_DEVICE" "$DEVICE_DELAY" &
    done
    
    # Esperar a que termine el lote
    wait
    
    batch_end=$(date +%s.%N)
    batch_duration=$(echo "$batch_end - $batch_start" | bc)
    BATCH_TIMES+=("$batch_duration")
    
    echo -e "${BLUE}✅ Lote $batch_number completado en ${batch_duration}s${NC}"
done

END_TIME=$(date +%s.%N)
TOTAL_DURATION=$(echo "$END_TIME - $TOTAL_START_TIME" | bc)
EXECUTION_DURATION=$(echo "$END_TIME - $START_TIME" | bc)

# Contar mensajes exitosos
success_count=$(ls /tmp/gobromq_success_*_$$ 2>/dev/null | wc -l)
total_expected=$((device_count * MESSAGES_PER_DEVICE))

# Calcular métricas
if [ $(echo "$EXECUTION_DURATION > 0" | bc) -eq 1 ]; then
    rate=$(echo "scale=2; $success_count / $EXECUTION_DURATION" | bc)
else
    rate="N/A"
fi

# Métricas de lotes
total_batch_time=0
fastest_batch_time=999999
slowest_batch_time=0
for batch_time in "${BATCH_TIMES[@]}"; do
    total_batch_time=$(echo "$total_batch_time + $batch_time" | bc)
    if [ $(echo "$batch_time < $fastest_batch_time" | bc) -eq 1 ]; then
        fastest_batch_time="$batch_time"
    fi
    if [ $(echo "$batch_time > $slowest_batch_time" | bc) -eq 1 ]; then
        slowest_batch_time="$batch_time"
    fi
done

if [ ${#BATCH_TIMES[@]} -gt 0 ]; then
    avg_batch_time=$(echo "scale=3; $total_batch_time / ${#BATCH_TIMES[@]}" | bc)
else
    avg_batch_time="N/A"
fi

# Métricas de mensajes individuales
if [ -f "/tmp/gobromq_msg_timing_$$" ]; then
    msg_count=$(wc -l < "/tmp/gobromq_msg_timing_$$")
    if [ $msg_count -gt 0 ]; then
        total_msg_time=$(awk -F: '{sum+=$3} END {print sum}' "/tmp/gobromq_msg_timing_$$")
        avg_msg_time=$(echo "scale=4; $total_msg_time / $msg_count" | bc)
        fastest_msg_time=$(awk -F: '{print $3}' "/tmp/gobromq_msg_timing_$$" | sort -n | head -1)
        slowest_msg_time=$(awk -F: '{print $3}' "/tmp/gobromq_msg_timing_$$" | sort -n | tail -1)
    else
        avg_msg_time="N/A"
        fastest_msg_time="N/A"
        slowest_msg_time="N/A"
    fi
else
    avg_msg_time="N/A"
    fastest_msg_time="N/A"
    slowest_msg_time="N/A"
fi

echo -e "\n${PURPLE}📈 MÉTRICAS DETALLADAS DE RENDIMIENTO${NC}"
echo "========================================"

echo -e "\n${WHITE}⏱️  TIEMPOS GENERALES:${NC}"
echo "------------------------"
echo -e "${GREEN}🕐 Tiempo total (incluye setup): ${TOTAL_DURATION}s${NC}"
echo -e "${GREEN}⚡ Tiempo de ejecución: ${EXECUTION_DURATION}s${NC}"
echo -e "${GREEN}🔧 Tiempo de configuración: $(echo "scale=2; $TOTAL_DURATION - $EXECUTION_DURATION" | bc)s${NC}"

echo -e "\n${WHITE}📊 ESTADÍSTICAS DE MENSAJES:${NC}"
echo "------------------------------"
echo -e "${GREEN}✅ Mensajes exitosos: ${success_count}/${total_expected}${NC}"
echo -e "${GREEN}🏢 Dispositivos simulados: ${device_count}${NC}"
echo -e "${GREEN}🏠 Zonas cubiertas: ${#ZONES[@]}${NC}"
echo -e "${GREEN}📱 Tipos de sensores: ${#DEVICE_TYPES[@]}${NC}"
echo -e "${GREEN}🚀 Throughput: ${rate} mensajes/segundo${NC}"

if [ $total_expected -gt 0 ]; then
    efficiency=$(echo "scale=1; $success_count * 100 / $total_expected" | bc)
    echo -e "${GREEN}📊 Eficiencia: ${efficiency}%${NC}"
fi

if [ $device_count -gt 0 ]; then
    avg_per_device=$(echo "scale=2; $success_count / $device_count" | bc)
    echo -e "${CYAN}📡 Promedio por dispositivo: ${avg_per_device} mensajes/dispositivo${NC}"
fi

echo -e "\n${WHITE}🔄 MÉTRICAS DE LOTES:${NC}"
echo "----------------------"
echo -e "${BLUE}📦 Total de lotes ejecutados: ${#BATCH_TIMES[@]}${NC}"
echo -e "${BLUE}⏱️  Tiempo promedio por lote: ${avg_batch_time}s${NC}"
echo -e "${BLUE}🏃 Lote más rápido: ${fastest_batch_time}s${NC}"
echo -e "${BLUE}🐌 Lote más lento: ${slowest_batch_time}s${NC}"
echo -e "${BLUE}📊 Dispositivos por lote: ${CONCURRENT_DEVICES}${NC}"

echo -e "\n${WHITE}📨 MÉTRICAS DE MENSAJES:${NC}"
echo "-------------------------"
echo -e "${YELLOW}⏱️  Tiempo promedio por mensaje: ${avg_msg_time}s${NC}"
echo -e "${YELLOW}🏃 Mensaje más rápido: ${fastest_msg_time}s${NC}"
echo -e "${YELLOW}🐌 Mensaje más lento: ${slowest_msg_time}s${NC}"

# Mostrar top 5 zonas con más dispositivos
echo -e "\n${WHITE}🏠 TOP 5 ZONAS CON MÁS DISPOSITIVOS:${NC}"
echo "------------------------------------"
zone_stats="/tmp/gobromq_zone_stats_$$"
> "$zone_stats"

for device_info in "${ALL_DEVICES[@]}"; do
    zone=$(echo "$device_info" | cut -d: -f2)
    echo "$zone" >> "$zone_stats"
done

if [ -f "$zone_stats" ]; then
    sort "$zone_stats" | uniq -c | sort -nr | head -5 | while read count zone; do
        echo -e "${CYAN}  📍 $zone: $count dispositivos${NC}"
    done
fi

# Limpiar archivos temporales
rm -f /tmp/gobromq_success_*_$$
rm -f /tmp/gobromq_timing_$$
rm -f /tmp/gobromq_msg_timing_$$
rm -f /tmp/gobromq_zone_stats_$$

echo -e "\n${GREEN}🎉 Simulación completada con métricas detalladas${NC}"
echo -e "${CYAN}📊 ${device_count} dispositivos, ${#ZONES[@]} zonas, ${success_count} mensajes exitosos${NC}"
