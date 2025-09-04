#!/bin/bash

# 🧪 GoBroMQ - Tests Simples y Efectivos
# ======================================

BROKER_HOST="localhost"
BROKER_PORT="1884"
USERNAME="admin"
PASSWORD="password123"

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Tests Simples de GoBroMQ${NC}"
echo "=============================="

# Verificar que mosquitto_pub está disponible
if ! command -v mosquitto_pub &> /dev/null; then
    echo -e "${RED}❌ mosquitto-clients no está instalado${NC}"
    echo "Instálalo con: sudo apt-get install mosquitto-clients"
    exit 1
fi

# Test 1: Verificar que el broker responde
echo -e "\n${BLUE}📡 Test 1: Conectividad${NC}"
echo "------------------------"
if timeout 5 bash -c "</dev/tcp/$BROKER_HOST/$BROKER_PORT" 2>/dev/null; then
    echo -e "${GREEN}✅ Puerto $BROKER_PORT está abierto${NC}"
else
    echo -e "${RED}❌ No se puede conectar al puerto $BROKER_PORT${NC}"
    echo "¿Está el broker ejecutándose? Ejecuta: docker-compose up -d"
    exit 1
fi

# Test 2: Publicar mensaje simple
echo -e "\n${BLUE}📤 Test 2: Publicación Simple${NC}"
echo "------------------------------"
if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
    -u $USERNAME -P $PASSWORD \
    -t 'test/simple' -m 'Hello from test' \
    -q 0 -i "simple_test_$$" 2>/dev/null; then
    echo -e "${GREEN}✅ Mensaje publicado exitosamente${NC}"
else
    echo -e "${RED}❌ Error al publicar mensaje${NC}"
    exit 1
fi

# Test 3: Publicar con QoS 1
echo -e "\n${BLUE}📤 Test 3: Publicación QoS 1${NC}"
echo "-----------------------------"
if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
    -u $USERNAME -P $PASSWORD \
    -t 'test/qos1' -m 'QoS 1 test message' \
    -q 1 -i "qos1_test_$$" 2>/dev/null; then
    echo -e "${GREEN}✅ Mensaje QoS 1 publicado exitosamente${NC}"
else
    echo -e "${RED}❌ Error al publicar mensaje QoS 1${NC}"
    exit 1
fi

# Test 4: Múltiples mensajes
echo -e "\n${BLUE}📤 Test 4: Múltiples Mensajes${NC}"
echo "-------------------------------"
success_count=0
total_messages=5

for i in $(seq 1 $total_messages); do
    if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
        -u $USERNAME -P $PASSWORD \
        -t "test/multi" -m "Mensaje $i" \
        -q 0 -i "multi_test_${i}_$$" 2>/dev/null; then
        ((success_count++))
        echo -e "${GREEN}✅ Mensaje $i enviado${NC}"
    else
        echo -e "${RED}❌ Error en mensaje $i${NC}"
    fi
done

echo -e "\n${BLUE}📊 Resultados: $success_count/$total_messages mensajes exitosos${NC}"

if [ $success_count -eq $total_messages ]; then
    echo -e "${GREEN}✅ TODOS LOS TESTS PASARON${NC}"
else
    echo -e "${YELLOW}⚠️  Algunos tests fallaron${NC}"
fi

# Test 5: Prueba de autenticación incorrecta
echo -e "\n${BLUE}🔐 Test 5: Autenticación Incorrecta${NC}"
echo "-----------------------------------"
if mosquitto_pub -V mqttv311 -h $BROKER_HOST -p $BROKER_PORT \
    -u "wrong_user" -P "wrong_pass" \
    -t 'test/auth' -m 'should fail' \
    -q 0 -i "auth_test_$$" 2>/dev/null; then
    echo -e "${RED}❌ Autenticación incorrecta fue aceptada (PROBLEMA)${NC}"
else
    echo -e "${GREEN}✅ Autenticación incorrecta rechazada correctamente${NC}"
fi

echo -e "\n${GREEN}🎉 Tests completados${NC}"
