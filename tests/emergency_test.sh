#!/bin/bash

# 🧪 GoBroMQ - Simulador de Alertas y Emergencias IoT
# ==================================================

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
BLINK='\033[5m'
NC='\033[0m'

echo -e "${RED}🚨 Simulador de Alertas y Emergencias IoT${NC}"
echo "=============================================="

# Dispositivos críticos para simulación de emergencias
declare -a CRITICAL_DEVICES=(
    "smoke_detector_living:sensors/safety/smoke/living:humo"
    "smoke_detector_kitchen:sensors/safety/smoke/kitchen:humo"
    "gas_detector_kitchen:sensors/safety/gas/kitchen:gas"
    "water_leak_bathroom:sensors/safety/water/bathroom:agua"
    "water_leak_kitchen:sensors/safety/water/kitchen:agua"
    "door_sensor_main:sensors/security/door/main:puerta"
    "window_sensor_living:sensors/security/window/living:ventana"
    "motion_detector_entrance:sensors/security/motion/entrance:movimiento"
    "panic_button_bedroom:sensors/security/panic/bedroom:panico"
    "camera_entrance:sensors/security/camera/entrance:camara"
    "temperature_server_room:sensors/critical/temp/server:temperatura"
    "power_monitor_main:sensors/critical/power/main:energia"
)

# Función para generar alertas de emergencia
generate_emergency_alert() {
    local device_info=$1
    local device_id=$(echo "$device_info" | cut -d: -f1)
    local topic=$(echo "$device_info" | cut -d: -f2)
    local sensor_type=$(echo "$device_info" | cut -d: -f3)
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    local alert_id="ALERT_$(date +%s)_$$"
    
    case $sensor_type in
        "humo")
            local smoke_level=$((70 + RANDOM % 30))  # Alto nivel de humo
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"FIRE_ALARM\",\"severity\":\"CRITICAL\",\"smoke_level\":${smoke_level},\"unit\":\"ppm\",\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Humo detectado - Posible incendio\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"activate_sprinklers\",\"sound_alarm\",\"call_fire_dept\"]}"
            ;;
        "gas")
            local gas_level=$((50 + RANDOM % 50))  # Nivel peligroso de gas
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"GAS_LEAK\",\"severity\":\"CRITICAL\",\"gas_level\":${gas_level},\"unit\":\"ppm\",\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Fuga de gas detectada\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"shut_gas_valve\",\"sound_alarm\",\"ventilate_area\"]}"
            ;;
        "agua")
            local water_detected=1
            local flow_rate=$((10 + RANDOM % 20))
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"WATER_LEAK\",\"severity\":\"HIGH\",\"water_detected\":${water_detected},\"flow_rate\":${flow_rate},\"unit\":\"L/min\",\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Fuga de agua detectada\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"shut_water_valve\",\"alert_maintenance\"]}"
            ;;
        "puerta"|"ventana")
            local forced_entry=1
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"SECURITY_BREACH\",\"severity\":\"HIGH\",\"forced_entry\":${forced_entry},\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Entrada forzada detectada\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"sound_alarm\",\"alert_security\",\"record_video\"]}"
            ;;
        "movimiento")
            local unauthorized_access=1
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"INTRUSION\",\"severity\":\"MEDIUM\",\"unauthorized_access\":${unauthorized_access},\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Movimiento no autorizado detectado\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"record_video\",\"alert_security\"]}"
            ;;
        "panico")
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"PANIC_BUTTON\",\"severity\":\"CRITICAL\",\"panic_activated\":1,\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Botón de pánico activado\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"call_emergency\",\"sound_alarm\",\"unlock_exits\"]}"
            ;;
        "camara")
            local face_recognition_failed=1
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"UNRECOGNIZED_PERSON\",\"severity\":\"MEDIUM\",\"face_recognition_failed\":${face_recognition_failed},\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Persona no reconocida detectada\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"record_video\",\"alert_security\"]}"
            ;;
        "temperatura")
            local temp=$((45 + RANDOM % 20))  # Temperatura crítica
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"OVERHEATING\",\"severity\":\"HIGH\",\"temperature\":${temp},\"unit\":\"celsius\",\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Temperatura crítica detectada\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"increase_cooling\",\"alert_maintenance\"]}"
            ;;
        "energia")
            local power_outage=1
            local backup_battery=$((20 + RANDOM % 30))
            echo "{\"alert_id\":\"${alert_id}\",\"device_id\":\"${device_id}\",\"type\":\"POWER_FAILURE\",\"severity\":\"HIGH\",\"power_outage\":${power_outage},\"backup_battery\":${backup_battery},\"unit\":\"percent\",\"location\":\"$(echo $topic | cut -d/ -f4)\",\"message\":\"Corte de energía detectado\",\"timestamp\":\"${timestamp}\",\"auto_actions\":[\"activate_backup\",\"alert_maintenance\"]}"
            ;;
    esac
}

# Función para simular escalamiento de emergencia
simulate_emergency_escalation() {
    local scenario=$1
    
    case $scenario in
        "fire")
            echo -e "${RED}${BLINK}🔥 SIMULANDO ESCENARIO DE INCENDIO${NC}"
            
            # Detector de humo inicial
            device_info="smoke_detector_kitchen:sensors/safety/smoke/kitchen:humo"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/critical/fire" -m "$alert_data" \
                -q 2 -i "fire_alert_1_$$" 2>/dev/null
            echo -e "${RED}🚨 Detector de humo en cocina activado${NC}"
            
            sleep 2
            
            # Detector de humo sala (propagación)
            device_info="smoke_detector_living:sensors/safety/smoke/living:humo"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/critical/fire" -m "$alert_data" \
                -q 2 -i "fire_alert_2_$$" 2>/dev/null
            echo -e "${RED}🚨 Fuego propagándose a la sala${NC}"
            
            sleep 1
            
            # Sistema de aspersores
            sprinkler_data="{\"device_id\":\"sprinkler_system\",\"type\":\"FIRE_SUPPRESSION\",\"action\":\"ACTIVATED\",\"zones\":[\"kitchen\",\"living_room\"],\"water_pressure\":45,\"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")\"}"
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "systems/fire_suppression" -m "$sprinkler_data" \
                -q 2 -i "sprinkler_$$" 2>/dev/null
            echo -e "${BLUE}💧 Sistema de aspersores activado${NC}"
            ;;
            
        "security")
            echo -e "${YELLOW}🛡️  SIMULANDO BRECHA DE SEGURIDAD${NC}"
            
            # Puerta principal forzada
            device_info="door_sensor_main:sensors/security/door/main:puerta"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/security/breach" -m "$alert_data" \
                -q 2 -i "security_alert_1_$$" 2>/dev/null
            echo -e "${YELLOW}🚪 Puerta principal forzada${NC}"
            
            sleep 1
            
            # Cámara de entrada detecta intruso
            device_info="camera_entrance:sensors/security/camera/entrance:camara"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/security/intrusion" -m "$alert_data" \
                -q 2 -i "security_alert_2_$$" 2>/dev/null
            echo -e "${YELLOW}📹 Cámara detecta persona no reconocida${NC}"
            
            sleep 1
            
            # Detector de movimiento
            device_info="motion_detector_entrance:sensors/security/motion/entrance:movimiento"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/security/intrusion" -m "$alert_data" \
                -q 2 -i "security_alert_3_$$" 2>/dev/null
            echo -e "${YELLOW}👤 Movimiento no autorizado confirmado${NC}"
            ;;
            
        "flood")
            echo -e "${CYAN}🌊 SIMULANDO INUNDACIÓN${NC}"
            
            # Fuga en baño
            device_info="water_leak_bathroom:sensors/safety/water/bathroom:agua"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/water/leak" -m "$alert_data" \
                -q 1 -i "water_alert_1_$$" 2>/dev/null
            echo -e "${CYAN}🚿 Fuga detectada en baño${NC}"
            
            sleep 2
            
            # Fuga en cocina (tubería principal)
            device_info="water_leak_kitchen:sensors/safety/water/kitchen:agua"
            alert_data=$(generate_emergency_alert "$device_info")
            mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
                -u $USERNAME -P $PASSWORD \
                -t "alerts/water/leak" -m "$alert_data" \
                -q 1 -i "water_alert_2_$$" 2>/dev/null
            echo -e "${CYAN}🔧 Tubería principal comprometida${NC}"
            ;;
    esac
}

echo -e "${WHITE}⚠️  Iniciando simulación de emergencias...${NC}"

# Test 1: Alertas individuales
echo -e "\n${BLUE}📊 Test 1: Alertas Individuales${NC}"
echo "================================"

alert_count=0
for device_info in "${CRITICAL_DEVICES[@]}"; do
    device_id=$(echo "$device_info" | cut -d: -f1)
    topic=$(echo "$device_info" | cut -d: -f2)
    
    # 30% de probabilidad de que cada sensor genere una alerta
    if [ $((RANDOM % 100)) -lt 30 ]; then
        alert_data=$(generate_emergency_alert "$device_info")
        
        if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
            -u $USERNAME -P $PASSWORD \
            -t "$topic" -m "$alert_data" \
            -q 1 -i "alert_${device_id}_$$" 2>/dev/null; then
            echo -e "${YELLOW}⚠️  ${device_id}: Alerta generada${NC}"
            ((alert_count++))
        fi
        
        sleep 0.3
    fi
done

echo -e "${GREEN}📊 Total de alertas generadas: ${alert_count}${NC}"

# Test 2: Escenarios de emergencia
echo -e "\n${BLUE}📊 Test 2: Escenarios de Emergencia${NC}"
echo "===================================="

scenarios=("fire" "security" "flood")
for scenario in "${scenarios[@]}"; do
    echo -e "\n${PURPLE}--- Escenario: ${scenario} ---${NC}"
    simulate_emergency_escalation "$scenario"
    sleep 3
done

# Test 3: Situación de pánico
echo -e "\n${BLUE}📊 Test 3: Activación de Pánico${NC}"
echo "==============================="

device_info="panic_button_bedroom:sensors/security/panic/bedroom:panico"
panic_data=$(generate_emergency_alert "$device_info")

mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
    -u $USERNAME -P $PASSWORD \
    -t "alerts/critical/panic" -m "$panic_data" \
    -q 2 -i "panic_alert_$$" 2>/dev/null

echo -e "${RED}${BLINK}🆘 BOTÓN DE PÁNICO ACTIVADO${NC}"

# Test 4: Estado del sistema post-emergencia
echo -e "\n${BLUE}📊 Test 4: Estado del Sistema${NC}"
echo "============================="

system_status="{\"system_id\":\"emergency_control\",\"status\":\"POST_EMERGENCY\",\"active_alerts\":$((alert_count + 4)),\"systems_activated\":[\"fire_suppression\",\"security_lockdown\",\"emergency_lighting\"],\"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")\",\"next_actions\":[\"assess_damage\",\"contact_authorities\",\"begin_recovery\"]}"

mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
    -u $USERNAME -P $PASSWORD \
    -t "system/emergency/status" -m "$system_status" \
    -q 2 -i "system_status_$$" 2>/dev/null

echo -e "${GREEN}📊 Estado del sistema reportado${NC}"

echo -e "\n${GREEN}🎉 Simulación de emergencias completada${NC}"
echo -e "${CYAN}💡 Se simularon múltiples escenarios de emergencia con alertas realistas${NC}"
echo -e "${CYAN}🚨 Los mensajes incluyen niveles de severidad y acciones automáticas${NC}"
