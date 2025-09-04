# 🧪 GoBroMQ - Suite de Tests IoT

Esta suite de tests está diseñada para probar exhaustivamente el broker MQTT GoBroMQ con escenarios realistas de IoT.

## 📋 Tests Disponibles

### 1. **Tests Básicos** (`simple_tests.sh`)
- ✅ Conectividad del broker
- ✅ Publicación simple (QoS 0)
- ✅ Publicación QoS 1
- ✅ Múltiples mensajes
- ✅ Autenticación incorrecta

**Uso:**
```bash
./tests/simple_tests.sh
```

### 2. **Test de Rendimiento** (`benchmark_test.sh`)
- 🚀 Throughput con sensores IoT
- ⚡ Latencia de sensores críticos
- 📊 Datos JSON realistas
- 🏠 Múltiples tipos de sensores

**Uso:**
```bash
./tests/benchmark_test.sh
```

### 3. **Tráfico Masivo IoT** (`massive_iot_test.sh`) ⭐
- 🏢 **140 dispositivos** IoT simulados
- 🏠 **20 zonas** diferentes (cocina, sala, dormitorios, etc.)
- 📱 **25 tipos** de sensores diferentes
- 🔄 **10 conexiones concurrentes**
- 📊 **700 mensajes** en total
- ⚡ Optimizado para alta carga

**Tipos de dispositivos simulados:**
- 🚪 Puertas y ventanas
- 👤 Detectores de movimiento
- 🌡️ Sensores de temperatura/humedad
- 💡 Luces e interruptores
- 🔥 Detectores de humo y gas
- 💧 Sensores de agua
- 📹 Cámaras de seguridad
- 🔊 Altavoces
- 🌬️ Ventiladores y AC
- 🔋 Monitores de energía
- 📊 Sensores de calidad del aire
- ☀️ Sensores UV
- 🎵 Sensores de sonido
- Y muchos más...

**Uso:**
```bash
./tests/massive_iot_test.sh
```

### 4. **Alertas y Emergencias** (`emergency_test.sh`)
- 🚨 **12 dispositivos críticos** de seguridad
- 🔥 Escenarios de **incendio**
- 🛡️ Brechas de **seguridad**
- 🌊 **Inundaciones**
- 🆘 Botones de **pánico**
- 📈 **Escalamiento** automático de emergencias

**Uso:**
```bash
./tests/emergency_test.sh
```

### 5. **Suite Interactiva** (`run_tests.sh`)
- 📋 Menú interactivo
- 🎯 Ejecutar tests individuales o todos
- 🛠️ Modo manual para comandos personalizados

**Uso:**
```bash
./tests/run_tests.sh
```

## 🚀 Ejecución Rápida

### Prerequisitos
```bash
# Instalar mosquitto clients
sudo apt-get install mosquitto-clients

# Iniciar el broker
docker-compose up -d gobromq
```

### Tests Recomendados

**Para desarrollo rápido:**
```bash
./tests/simple_tests.sh
```

**Para probar rendimiento:**
```bash
./tests/massive_iot_test.sh
```

**Para probar emergencias:**
```bash
./tests/emergency_test.sh
```

**Suite completa:**
```bash
./tests/run_tests.sh
# Seleccionar opción 4 (Ejecutar todos los tests)
```

## 📊 Ejemplos de Datos JSON

### Sensor Normal
```json
{
  "device_id": "living_room_temperature_01",
  "zone": "sala",
  "type": "temperature",
  "temperature": 22,
  "target": 24,
  "unit": "celsius",
  "battery": 85,
  "signal": -45,
  "timestamp": "2025-09-03T10:30:45.123Z"
}
```

### Alerta de Emergencia
```json
{
  "alert_id": "ALERT_1725360645_123",
  "device_id": "smoke_detector_kitchen",
  "type": "FIRE_ALARM",
  "severity": "CRITICAL",
  "smoke_level": 85,
  "unit": "ppm",
  "location": "kitchen",
  "message": "Humo detectado - Posible incendio",
  "timestamp": "2025-09-03T10:30:45.123Z",
  "auto_actions": ["activate_sprinklers", "sound_alarm", "call_fire_dept"]
}
```

## 🏗️ Arquitectura de Topics

### Dispositivos Normales
```
iot/{zona}/{tipo_dispositivo}/{device_id}
```

**Ejemplos:**
- `iot/cocina/temperature/kitchen_temperature_01`
- `iot/sala/motion/living_room_motion_01`
- `iot/dormitorio_1/light/bedroom_1_light_01`

### Alertas y Emergencias
```
alerts/{categoria}/{tipo}
sensors/safety/{tipo}/{ubicacion}
sensors/security/{tipo}/{ubicacion}
```

**Ejemplos:**
- `alerts/critical/fire`
- `alerts/security/breach`
- `sensors/safety/smoke/kitchen`
- `sensors/security/motion/entrance`

## 📈 Métricas de Rendimiento

Con el test de tráfico masivo puedes esperar:

- **📊 140 dispositivos** simulados concurrentemente
- **🚀 ~50-100 mensajes/segundo** (dependiendo del hardware)
- **⚡ <100ms** de latencia promedio
- **✅ >95%** de éxito en entrega de mensajes
- **🔄 10 conexiones** MQTT concurrentes

## 🔧 Configuración

### Variables de Entorno
```bash
BROKER_HOST="localhost"     # Host del broker
BROKER_PORT="1884"          # Puerto del broker
USERNAME="admin"            # Usuario MQTT
PASSWORD="password123"      # Contraseña MQTT
```

### Personalización
Puedes modificar los tests editando:
- **Número de dispositivos**: Variable `device_count` en cada test
- **Frecuencia de mensajes**: Variable `DEVICE_DELAY`
- **Conexiones concurrentes**: Variable `CONCURRENT_DEVICES`
- **Tipos de sensores**: Array `DEVICE_TYPES`
- **Zonas**: Array `ZONES`

## 🐛 Troubleshooting

### Broker no responde
```bash
# Verificar que el broker está ejecutándose
docker-compose ps

# Ver logs del broker
docker-compose logs gobromq

# Reiniciar el broker
docker-compose restart gobromq
```

### Tests fallan
```bash
# Verificar mosquitto-clients
mosquitto_pub --help

# Test manual
mosquitto_pub -V mqttv311 -h localhost -p 1884 -u admin -P password123 -t 'test/manual' -m 'test message' -q 0 -i 'manual_test'
```

### Problemas de permisos
```bash
# Dar permisos a todos los tests
chmod +x tests/*.sh
```

## 🎯 Casos de Uso

### Desarrollo
- Usar `simple_tests.sh` para verificación rápida
- Usar `benchmark_test.sh` para medir rendimiento básico

### Testing de Carga
- Usar `massive_iot_test.sh` para probar con alta carga
- Incrementar `CONCURRENT_DEVICES` para más presión

### Testing de Escenarios Críticos
- Usar `emergency_test.sh` para probar manejo de alertas
- Verificar QoS 2 para mensajes críticos

### CI/CD
```bash
# Ejecutar todos los tests automáticamente
./tests/simple_tests.sh && ./tests/massive_iot_test.sh && ./tests/emergency_test.sh
```

---

**🎉 ¡Con esta suite puedes probar tu broker MQTT con escenarios muy realistas de IoT!**
