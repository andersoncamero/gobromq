#!/bin/bash

# 🧪 GoBroMQ - Suite de Tests Funcionales
# =======================================

BROKER_HOST="localhost"
BROKER_PORT="1884"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${PURPLE}🚀 GoBroMQ - Suite de Tests Funcionales${NC}"
echo "==========================================="

# Verificar que el broker está ejecutándose
if ! timeout 3 bash -c "</dev/tcp/$BROKER_HOST/$BROKER_PORT" 2>/dev/null; then
    echo -e "${RED}❌ Broker no accesible en $BROKER_HOST:$BROKER_PORT${NC}"
    echo -e "${YELLOW}💡 Para iniciar el broker ejecuta: docker-compose up -d${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Broker accesible en $BROKER_HOST:$BROKER_PORT${NC}"

# Función para mostrar menú
show_menu() {
    echo -e "\n${CYAN}🧪 Selecciona el tipo de test:${NC}"
    echo "=============================="
    echo -e "${BLUE}1)${NC} Tests Básicos (rápido)"
    echo -e "${BLUE}2)${NC} Test de Benchmark (rendimiento)"
    echo -e "${BLUE}3)${NC} Test de Tráfico Masivo IoT (realista)"
    echo -e "${BLUE}4)${NC} Test Masivo con Métricas Detalladas ⭐"
    echo -e "${BLUE}5)${NC} Ejecutar todos los tests"
    echo -e "${BLUE}6)${NC} Test manual interactivo"
    echo -e "${BLUE}7)${NC} Salir"
    echo ""
    echo -n "Opción [1-7]: "
}

# Test manual interactivo
interactive_test() {
    echo -e "\n${PURPLE}🎮 Modo Test Interactivo${NC}"
    echo "========================"
    echo -e "${YELLOW}💡 Puedes escribir comandos para probar el broker${NC}"
    echo -e "${YELLOW}💡 Ejemplos de comandos útiles:${NC}"
    echo -e "   ${GREEN}mosquitto_pub -h $BROKER_HOST -p $BROKER_PORT -V mqttv311 -t test -m 'hola'${NC}"
    echo -e "   ${GREEN}mosquitto_sub -h $BROKER_HOST -p $BROKER_PORT -V mqttv311 -t test${NC}"
    echo -e "   ${GREEN}ps aux | grep mosquitto${NC}"
    echo -e "   ${GREEN}exit${NC} para salir"
    echo
    
    while true; do
        echo -n -e "${GREEN}GoBroMQ-Test>${NC} "
        read -r command
        if [ -z "$command" ]; then
            continue
        fi
        if [ "$command" = "exit" ] || [ "$command" = "quit" ]; then
            break
        fi
        eval "$command"
    done
}

# Bucle principal
while true; do
    show_menu
    read -r choice
    
    case $choice in
        1)
            echo -e "\n${BLUE}🧪 Ejecutando Tests Básicos...${NC}"
            echo "================================"
            if [ -f "./tests/simple_tests.sh" ]; then
                ./tests/simple_tests.sh
            else
                echo -e "${RED}❌ Archivo simple_tests.sh no encontrado${NC}"
            fi
            ;;
        2)
            echo -e "\n${BLUE}🚀 Ejecutando Test de Benchmark...${NC}"
            echo "=================================="
            if [ -f "./tests/benchmark_test.sh" ]; then
                ./tests/benchmark_test.sh
            else
                echo -e "${RED}❌ Archivo benchmark_test.sh no encontrado${NC}"
            fi
            ;;
        3)
            echo -e "\n${BLUE}🌊 Ejecutando Test de Tráfico Masivo IoT...${NC}"
            echo "============================================"
            if [ -f "./tests/massive_iot_test.sh" ]; then
                ./tests/massive_iot_test.sh
            else
                echo -e "${RED}❌ Archivo massive_iot_test.sh no encontrado${NC}"
            fi
            ;;
        4)
            echo -e "\n${BLUE}📊 Ejecutando Test Masivo con Métricas Detalladas...${NC}"
            echo "==================================================="
            if [ -f "./tests/massive_iot_test_metrics.sh" ]; then
                ./tests/massive_iot_test_metrics.sh
            else
                echo -e "${RED}❌ Archivo massive_iot_test_metrics.sh no encontrado${NC}"
            fi
            ;;
        5)
            echo -e "\n${BLUE}🎯 Ejecutando TODOS los tests...${NC}"
            echo "================================"
            
            echo -e "\n${PURPLE}>>> Tests Básicos <<<${NC}"
            if [ -f "./tests/simple_tests.sh" ]; then
                ./tests/simple_tests.sh
            fi
            
            echo -e "\n${PURPLE}>>> Tests de Rendimiento <<<${NC}"
            if [ -f "./tests/benchmark_test.sh" ]; then
                ./tests/benchmark_test.sh
            fi
            
            echo -e "\n${PURPLE}>>> Test de Tráfico Masivo IoT <<<${NC}"
            if [ -f "./tests/massive_iot_test.sh" ]; then
                ./tests/massive_iot_test.sh
            fi
            
            echo -e "\n${GREEN}🎉 Todos los tests completados${NC}"
            ;;
        6)
            interactive_test
            ;;
        7)
            echo -e "\n${GREEN}👋 ¡Gracias por usar GoBroMQ!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Opción inválida. Por favor selecciona 1-7.${NC}"
            ;;
    esac
    
    echo -e "\n${YELLOW}Presiona Enter para continuar...${NC}"
    read -r
done
