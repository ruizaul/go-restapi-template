# 🗄️ TacoShare Delivery Database Schema

**Base de Datos:** PostgreSQL 18  
**Nombre:** `tacoshare_delivery`  
**Conexión:** Ver archivo `.env` para cadena de conexión (DATABASE_URL)

---

## 📋 Índice de Tablas

1. [**users**](#1-users) - Usuarios del sistema (clientes, comerciantes, conductores, admins)
2. [**merchants**](#2-merchants) - Información de negocios/tiendas
3. [**orders**](#3-orders) - Órdenes de delivery
4. [**driver_locations**](#4-driver_locations) - Ubicación en tiempo real de conductores
5. [**order_assignments**](#5-order_assignments) - Historial de asignaciones de órdenes a conductores
6. [**user_documents**](#6-user_documents) - Documentos de verificación de usuarios (KYC)
7. [**notifications**](#7-notifications) - Notificaciones push para usuarios
8. [**fcm_tokens**](#8-fcm_tokens) - Tokens FCM para notificaciones push
9. [**refresh_tokens**](#9-refresh_tokens) - Tokens de refresco JWT
10. [**schema_migrations**](#10-schema_migrations) - Control de versiones de migraciones

---

## 1. users

**Descripción:** Tabla principal de usuarios que soporta múltiples roles (customer, merchant, driver, admin). Implementa autenticación dual: email/password y OTP por teléfono con seguridad mejorada (hash SHA-256 + rate limiting).

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `name` | varchar(255) | NOT NULL | Nombre completo del usuario |
| `email` | varchar(255) | UNIQUE, nullable | Email (nullable para auth por OTP) |
| `phone` | varchar(20) | UNIQUE, NOT NULL | Teléfono (E.164 format) |
| `phone_encrypted` | bytea | nullable | 🔒 Teléfono cifrado (AES-256, clave en KMS) |
| `password_hash` | text | nullable | Hash bcrypt (nullable para auth por OTP) |
| `role` | varchar(50) | NOT NULL | Rol: customer, merchant, driver, admin |
| `created_at` | timestamptz | DEFAULT NOW() | Fecha de creación |
| `updated_at` | timestamptz | DEFAULT NOW() | Fecha de actualización |
| `otp_code` | varchar(6) | nullable | ⚠️ DEPRECATED: Código OTP texto plano (usar otp_hash) |
| `otp_hash` | varchar(64) | nullable | 🔒 SHA-256 hash de OTP + pepper del servidor |
| `otp_expires_at` | timestamptz | nullable | Expiración del OTP (10 minutos) |
| `otp_attempts` | integer | DEFAULT 0 | 🔒 Contador de intentos fallidos de OTP |
| `otp_locked_until` | timestamptz | nullable | 🔒 Lockout temporal (15 min tras 3 intentos) |
| `phone_verified` | boolean | DEFAULT false | Si el teléfono está verificado |
| `account_status` | varchar(20) | DEFAULT 'pending' | Estado: pending, active, suspended |
| `deleted_at` | timestamptz | nullable | 🔒 Soft delete para compliance (NULL = activo) |
| `first_name` | varchar(100) | nullable | Nombre(s) |
| `last_name` | varchar(100) | nullable | Apellido paterno |
| `mother_last_name` | varchar(100) | nullable | Apellido materno |
| `birth_date` | date | nullable | Fecha de nacimiento |

### Índices

- `users_pkey`: PRIMARY KEY (id)
- `users_email_key`: UNIQUE (email)
- `users_phone_key`: UNIQUE (phone)
- `idx_users_email`: (email)
- `idx_users_role`: (role)
- `idx_users_account_status`: (account_status)
- `idx_users_phone_otp`: (phone, otp_code) WHERE otp_code IS NOT NULL (DEPRECATED)
- `idx_users_otp_hash`: 🔒 (phone, otp_hash) WHERE otp_hash IS NOT NULL
- `idx_users_otp_locked`: 🔒 (otp_locked_until) WHERE otp_locked_until IS NOT NULL
- `idx_users_deleted_at`: 🔒 (deleted_at) WHERE deleted_at IS NULL

### Constraints

- **CHECK `users_role_check`**: role IN ('customer', 'merchant', 'driver', 'admin')
- **CHECK `users_account_status_check`**: account_status IN ('pending', 'active', 'suspended')
- **CHECK `check_active_user_credentials`**: Si account_status = 'active', email y password_hash deben estar presentes

### Relaciones

**Referenced by:**
- `merchants.user_id` → CASCADE DELETE
- `driver_locations.driver_id` → CASCADE DELETE
- `order_assignments.driver_id` → CASCADE DELETE
- `orders.driver_id` → SET NULL DELETE
- `orders.cancelled_by` → SET NULL DELETE
- `user_documents.user_id` → CASCADE DELETE
- `notifications.user_id` → CASCADE DELETE
- `fcm_tokens.user_id` → CASCADE DELETE
- `refresh_tokens.user_id` → CASCADE DELETE

### Flujo de Autenticación Seguro

1. **Registro por OTP (Seguro):**
   - Usuario envía phone → genera OTP de 6 dígitos (crypto/rand)
   - **Almacena SHA-256(OTP + OTP_PEPPER)** en `otp_hash` (NUNCA texto plano)
   - Almacena `otp_expires_at` (10 minutos)
   - Resetea `otp_attempts = 0`, `otp_locked_until = NULL`
   - Envía OTP por SMS (Twilio)

2. **Verificación OTP con Rate Limiting:**
   - Usuario envía phone + OTP → verifica si `otp_locked_until > NOW()` (lockout activo)
   - Compara SHA-256(OTP ingresado + OTP_PEPPER) con `otp_hash` almacenado
   - **Si falla**: incrementa `otp_attempts`, si >= 3 → `otp_locked_until = NOW() + 15 min`
   - **Si éxito**: marca `phone_verified = true`, limpia `otp_hash`, resetea `otp_attempts`

3. **Login por Email/Password:**
   - Usuario envía email + password → valida bcrypt hash
   - Retorna JWT access token (15 min) + refresh token (90 días)

### Seguridad OTP

- ✅ **Hash SHA-256 + Pepper**: OTP nunca almacenado en texto plano
- ✅ **Rate Limiting**: 3 intentos máximo, lockout de 15 minutos
- ✅ **TTL**: OTP expira en 10 minutos
- ✅ **Borrado inmediato**: OTP hash eliminado tras verificación exitosa
- ✅ **Pepper del servidor**: Variable `OTP_PEPPER` en .env (no en BD)

---

## 2. merchants

**Descripción:** Información de negocios/tiendas que generan órdenes de delivery. Cada merchant está vinculado a un usuario con rol 'merchant'.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `user_id` | uuid | FK users(id), NOT NULL, UNIQUE | Usuario propietario |
| `business_name` | varchar(255) | NOT NULL | Nombre del negocio |
| `business_type` | varchar(100) | NOT NULL | Tipo: restaurant, grocery, pharmacy, etc. |
| `phone` | varchar(20) | NOT NULL | Teléfono de contacto |
| `email` | varchar(255) | nullable | Email de contacto |
| `address` | text | NOT NULL | Dirección completa |
| `latitude` | numeric(10,8) | NOT NULL | Latitud de pickup |
| `longitude` | numeric(11,8) | NOT NULL | Longitud de pickup |
| `city` | varchar(100) | NOT NULL | Ciudad |
| `state` | varchar(100) | NOT NULL | Estado |
| `postal_code` | varchar(10) | nullable | Código postal |
| `country` | varchar(2) | DEFAULT 'MX' | Código país ISO-2 |
| `status` | varchar(20) | DEFAULT 'active', NOT NULL | Estado: active, inactive, suspended |
| `rating` | numeric(3,2) | DEFAULT 0.00 | Calificación 0.00-5.00 |
| `total_orders` | integer | DEFAULT 0 | Total de órdenes completadas |
| `created_at` | timestamptz | DEFAULT NOW() | Fecha de creación |
| `updated_at` | timestamptz | DEFAULT NOW() | Fecha de actualización |

### Índices

- `merchants_pkey`: PRIMARY KEY (id)
- `idx_merchants_user_id_unique`: UNIQUE (user_id)
- `idx_merchants_user_id`: (user_id)
- `idx_merchants_status`: (status)
- `idx_merchants_city`: (city)
- `idx_merchants_business_type`: (business_type)
- `idx_merchants_location`: (latitude, longitude) — para búsquedas geoespaciales

### Constraints

- **CHECK `merchants_status_check`**: status IN ('active', 'inactive', 'suspended')
- **CHECK `merchants_rating_check`**: rating >= 0 AND rating <= 5

### Relaciones

**References:**
- `user_id` → `users(id)` RESTRICT DELETE 🔒

### Row-Level Security (RLS)

**Políticas habilitadas:**
- ✅ **SELECT**: Solo el propio usuario puede ver sus documentos
- ✅ **INSERT**: Solo el propio usuario puede crear sus documentos
- ✅ **UPDATE**: Solo el propio usuario puede actualizar sus documentos
- ✅ **DELETE**: Solo el propio usuario puede eliminar sus documentos
- ✅ **Admin override**: Rol admin puede ver todos los documentos

### Encriptación de PII

**Columnas cifradas (AES-256):**
- `fiscal_rfc`: RFC cifrado
- `fiscal_certificate_url`: URL cifrada (documentos sensibles)

**Helpers de encriptación:**
```sql
-- Cifrar
UPDATE user_documents SET fiscal_rfc = encrypt_text('XAXX010101000');

-- Descifrar
SELECT decrypt_text(fiscal_rfc) FROM user_documents;
```

**Referenced by:**
- `orders.merchant_id` → RESTRICT DELETE (no puede borrarse si tiene órdenes)

### Uso en Sistema de Asignación

- Las coordenadas `(latitude, longitude)` se usan como punto de **pickup** en órdenes
- El `business_type` permite filtrado y categorización
- `status = 'active'` determina si puede recibir nuevas órdenes

---

## 3. orders

**Descripción:** Órdenes de delivery con máquina de estados completa. Contiene toda la información de pickup, delivery, items, y seguimiento temporal.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `external_order_id` | varchar(255) | nullable | ID de orden del backend externo |
| `merchant_id` | uuid | FK merchants(id), NOT NULL | Negocio origen |
| `driver_id` | uuid | FK users(id), nullable | Conductor asignado |
| `customer_name` | varchar(255) | NOT NULL | Nombre del cliente |
| `customer_phone` | varchar(20) | NOT NULL | Teléfono del cliente |
| `pickup_address` | text | NOT NULL | Dirección de recogida |
| `pickup_latitude` | numeric(10,8) | NOT NULL | Latitud de pickup |
| `pickup_longitude` | numeric(11,8) | NOT NULL | Longitud de pickup |
| `pickup_instructions` | text | nullable | Instrucciones de recogida |
| `delivery_address` | text | NOT NULL | Dirección de entrega |
| `delivery_latitude` | numeric(10,8) | NOT NULL | Latitud de entrega |
| `delivery_longitude` | numeric(11,8) | NOT NULL | Longitud de entrega |
| `delivery_instructions` | text | nullable | Instrucciones de entrega |
| `items` | jsonb | NOT NULL | Array JSON de items {name, quantity, price} |
| `total_amount` | numeric(10,2) | NOT NULL, > 0 | Monto total de la orden |
| `delivery_fee` | numeric(10,2) | DEFAULT 0.00 | Tarifa de delivery |
| `status` | varchar(50) | DEFAULT 'searching_driver', NOT NULL | Estado actual de la orden |
| `distance_km` | numeric(6,2) | nullable | Distancia total pickup→delivery |
| `estimated_duration_minutes` | integer | nullable | Duración estimada |
| `created_at` | timestamptz | DEFAULT NOW() | Fecha de creación |
| `updated_at` | timestamptz | DEFAULT NOW() | Última actualización |
| `assigned_at` | timestamptz | nullable | Cuando se asignó conductor |
| `accepted_at` | timestamptz | nullable | Cuando conductor aceptó |
| `picked_up_at` | timestamptz | nullable | Cuando recogió la orden |
| `delivered_at` | timestamptz | nullable | Cuando se entregó |
| `cancelled_at` | timestamptz | nullable | Cuando se canceló |
| `cancellation_reason` | text | nullable | Razón de cancelación |
| `cancelled_by` | uuid | FK users(id), nullable | Usuario que canceló |
| `delivery_code` | varchar(4) | NOT NULL | 🔒 Código criptográfico de 4 dígitos (crypto/rand, sin default) |
| `delivery_code_attempts` | integer | DEFAULT 0 | 🔒 Intentos fallidos de verificación (máx 3) |
| `customer_phone_encrypted` | bytea | nullable | 🔒 Teléfono cifrado (AES-256, clave en KMS) |
| `pickup_address_encrypted` | bytea | nullable | 🔒 Dirección pickup cifrada (AES-256) |
| `delivery_address_encrypted` | bytea | nullable | 🔒 Dirección delivery cifrada (AES-256) |

### Índices

- `orders_pkey`: PRIMARY KEY (id)
- `idx_orders_merchant_id`: (merchant_id)
- `idx_orders_driver_id`: (driver_id)
- `idx_orders_status`: (status)
- `idx_orders_created_at`: (created_at DESC)
- `idx_orders_external_id`: (external_order_id)
- `idx_orders_pickup_location`: (pickup_latitude, pickup_longitude)
- `idx_orders_delivery_location`: (delivery_latitude, delivery_longitude)
- `idx_orders_driver_status`: (driver_id, status) WHERE status IN ('assigned', 'accepted', 'picked_up', 'in_transit')
- `idx_orders_delivery_code`: (id, delivery_code)

### Constraints

- **CHECK `orders_status_check`**: status IN ('searching_driver', 'assigned', 'accepted', 'picked_up', 'in_transit', 'delivered', 'cancelled', 'no_driver_available')
- **CHECK `orders_total_amount_check`**: total_amount > 0
- **CHECK `check_delivery_code_format`**: delivery_code ~ '^\d{4}$' (4 dígitos numéricos)

### Relaciones

**References:**
- `merchant_id` → `merchants(id)` RESTRICT DELETE
- `driver_id` → `users(id)` SET NULL DELETE
- `cancelled_by` → `users(id)` SET NULL DELETE

**Referenced by:**
- `order_assignments.order_id` → CASCADE DELETE

### Máquina de Estados

```
searching_driver → assigned → accepted → picked_up → in_transit → delivered
       ↓              ↓           ↓            ↓          ↓
   cancelled      cancelled   cancelled   cancelled  cancelled
       ↓
no_driver_available (si no hay conductores tras reintentos)
```

**Estados:**
1. **searching_driver**: Orden creada, buscando conductor disponible
2. **assigned**: Conductor asignado, esperando aceptación (timeout: 15s)
3. **accepted**: Conductor aceptó, va hacia pickup
4. **picked_up**: Conductor recogió la orden
5. **in_transit**: Conductor en camino a delivery
6. **delivered**: Orden entregada exitosamente (requiere `delivery_code`)
7. **cancelled**: Orden cancelada (por merchant, driver, o admin)
8. **no_driver_available**: No se encontró conductor después de reintentos

### Campos Temporales

| Campo | Se llena cuando... |
|-------|-------------------|
| `created_at` | Se crea la orden |
| `assigned_at` | Status → assigned |
| `accepted_at` | Status → accepted |
| `picked_up_at` | Status → picked_up |
| `delivered_at` | Status → delivered |
| `cancelled_at` | Status → cancelled |

### Delivery Code Flow (Seguro)

1. **Generación criptográfica:**
   - Al crear orden → genera código con `crypto/rand` (NO math/rand)
   - 4 dígitos numéricos únicos (0000-9999)
   - Se muestra al customer en la app

2. **Verificación con Rate Limiting:**
   - Driver ingresa código para marcar como `delivered`
   - Validación: `order_id` + `delivery_code` deben coincidir
   - **Máximo 3 intentos** (`delivery_code_attempts`)
   - Cada intento fallido incrementa contador
   - Si >= 3 intentos → orden bloqueada, requiere intervención admin

3. **Auditoría:**
   - Todos los intentos (éxito/fallo) se registran en `delivery_code_audit`
   - Incluye: code ingresado, IP, user_agent, timestamp
   - Permite detectar patrones de ataque por fuerza bruta

### Seguridad del Delivery Code

- ✅ **Generación criptográfica**: `crypto/rand.Int()` (no predecible)
- ✅ **Rate limiting**: 3 intentos máximo
- ✅ **Auditoría completa**: Todos los intentos registrados
- ✅ **Sin default inseguro**: Removido `DEFAULT '0000'`
- ✅ **Validación de formato**: CHECK constraint regex `^\d{4}$`

---

## 4. driver_locations

**Descripción:** Ubicación GPS en tiempo real de conductores. Actualizada cada 5-10 segundos por la app del driver. Esencial para el algoritmo de asignación de órdenes.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `driver_id` | uuid | FK users(id), UNIQUE, NOT NULL | Conductor (rol driver) |
| `latitude` | numeric(10,8) | NOT NULL | Latitud actual |
| `longitude` | numeric(11,8) | NOT NULL | Longitud actual |
| `heading` | numeric(5,2) | nullable | Dirección de movimiento (0-360°) |
| `speed_kmh` | numeric(5,2) | nullable | Velocidad en km/h |
| `accuracy_meters` | numeric(6,2) | nullable | Precisión GPS en metros |
| `is_available` | boolean | DEFAULT true | Si está disponible para recibir órdenes |
| `updated_at` | timestamptz | DEFAULT NOW() | Última actualización GPS |

### Índices

- `driver_locations_pkey`: PRIMARY KEY (id)
- `idx_driver_locations_driver_unique`: UNIQUE (driver_id)
- `idx_driver_locations_driver_id`: (driver_id)
- `idx_driver_locations_location`: (latitude, longitude)
- `idx_driver_locations_available`: (is_available) WHERE is_available = true
- `idx_driver_locations_available_location`: (is_available, latitude, longitude) WHERE is_available = true
- `idx_driver_locations_updated_at`: (updated_at DESC)

### Relaciones

**References:**
- `driver_id` → `users(id)` RESTRICT DELETE 🔒

### Uso en Algoritmo de Asignación

**Proceso de búsqueda de conductores:**

1. **Query geoespacial:**
   ```sql
   SELECT driver_id, latitude, longitude
   FROM driver_locations
   WHERE is_available = true
   AND updated_at > NOW() - INTERVAL '5 minutes'  -- Ubicación reciente
   ORDER BY distance_from_pickup ASC  -- Cálculo Haversine
   LIMIT 10
   ```

2. **Radio de búsqueda incremental:**
   - Intento 1: 2km de radio (`ASSIGNMENT_RADIUS_KM`)
   - Intento 2: 5km de radio
   - Intento 3: 10km de radio
   - Si no hay conductores → `no_driver_available`

3. **Disponibilidad:**
   - `is_available = true`: conductor activo y sin orden actual
   - `is_available = false`: conductor ocupado, offline, o pausado

4. **Timeout de ubicación:**
   - Si `updated_at` > 5 minutos → conductor considerado offline

### Campos GPS Adicionales

- **`heading`**: Útil para predecir dirección de movimiento
- **`speed_kmh`**: Detectar si está en movimiento o estacionado
- **`accuracy_meters`**: Filtrar ubicaciones imprecisas (> 100m)

---

## 5. order_assignments

**Descripción:** Historial de intentos de asignación de órdenes a conductores. Cada fila representa un intento de asignar una orden específica a un conductor específico.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `order_id` | uuid | FK orders(id), NOT NULL | Orden a asignar |
| `driver_id` | uuid | FK users(id), NOT NULL | Conductor al que se ofreció |
| `attempt_number` | integer | NOT NULL | Número de intento secuencial (1, 2, 3...) |
| `search_radius_km` | numeric(6,2) | NOT NULL | Radio de búsqueda usado (2km, 5km, 10km) |
| `distance_to_pickup_km` | numeric(6,2) | NOT NULL | Distancia driver → pickup |
| `estimated_arrival_minutes` | integer | nullable | Tiempo estimado de llegada |
| `status` | varchar(20) | DEFAULT 'pending', NOT NULL | Estado: pending, accepted, rejected, timeout, expired |
| `created_at` | timestamptz | DEFAULT NOW() | Momento de la asignación |
| `responded_at` | timestamptz | nullable | Cuando el conductor respondió |
| `expires_at` | timestamptz | NOT NULL | Expiración del timeout (created_at + 15s) |
| `rejection_reason` | text | nullable | Razón si rechazó |

### Índices

- `order_assignments_pkey`: PRIMARY KEY (id)
- `idx_order_assignments_order_id`: (order_id)
- `idx_order_assignments_driver_id`: (driver_id)
- `idx_order_assignments_status`: (status)
- `idx_order_assignments_expires_at`: (expires_at)
- `idx_order_assignments_created_at`: (created_at DESC)
- `idx_order_assignments_pending`: (order_id, status, expires_at) WHERE status = 'pending'

### Constraints

- **CHECK `order_assignments_status_check`**: status IN ('pending', 'accepted', 'rejected', 'timeout', 'expired')

### Relaciones

**References:**
- `order_id` → `orders(id)` CASCADE DELETE
- `driver_id` → `users(id)` CASCADE DELETE

### Algoritmo de Asignación (Round-Robin)

**Variables de configuración (.env):**
- `ASSIGNMENT_RADIUS_KM = 2.0` — Radio inicial de búsqueda
- `ASSIGNMENT_TIMEOUT_SECONDS = 15` — Timeout por conductor
- `ASSIGNMENT_RETRY_INTERVAL_SECONDS = 15` — Pausa entre reintentos
- `ASSIGNMENT_MAX_SEARCH_SECONDS = 180` — Tiempo máximo total (3 min)

**Flujo de asignación:**

1. **Buscar conductores en radio de 2km:**
   ```sql
   -- Ordenados por distancia (más cercano primero)
   SELECT driver_id FROM driver_locations
   WHERE is_available = true
   AND distance_to_pickup <= 2.0
   ORDER BY distance_to_pickup ASC
   ```

2. **Por cada conductor:**
   - Crear registro `order_assignments`:
     - `attempt_number` = siguiente número
     - `search_radius_km` = 2.0
     - `distance_to_pickup_km` = distancia calculada
     - `status` = 'pending'
     - `expires_at` = NOW() + 15 segundos
   - Enviar notificación push al conductor
   - Esperar 15 segundos

3. **Resultado por conductor:**
   - **Aceptó:** `status = 'accepted'`, `orders.status = 'assigned'`, `orders.driver_id = driver_id`
   - **Rechazó:** `status = 'rejected'`, continuar con siguiente conductor
   - **Timeout:** `status = 'timeout'`, continuar con siguiente conductor
   - **Expiró:** `status = 'expired'` (si pasó el tiempo mientras se procesaba)

4. **Si ningún conductor aceptó en radio de 2km:**
   - Pausar 15 segundos (`ASSIGNMENT_RETRY_INTERVAL_SECONDS`)
   - Repetir con radio de 5km
   - Si no hay conductores → pausar 15s y repetir con 10km

5. **Si no hay conductor después de 3 minutos:**
   - `orders.status = 'no_driver_available'`
   - Notificar al merchant

### Análisis de Datos

**Métricas calculables desde esta tabla:**

```sql
-- Tasa de aceptación por conductor
SELECT driver_id,
  COUNT(*) FILTER (WHERE status = 'accepted') * 100.0 / COUNT(*) AS acceptance_rate
FROM order_assignments
GROUP BY driver_id;

-- Promedio de intentos antes de asignar
SELECT AVG(attempt_number) FROM order_assignments WHERE status = 'accepted';

-- Tiempo promedio de respuesta
SELECT AVG(EXTRACT(EPOCH FROM (responded_at - created_at))) AS avg_response_seconds
FROM order_assignments
WHERE status IN ('accepted', 'rejected');
```

---

## 6. user_documents

**Descripción:** Documentos de verificación KYC (Know Your Customer) para drivers y merchants. Almacena URLs de documentos subidos a Cloudflare R2.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `user_id` | uuid | FK users(id), UNIQUE, NOT NULL | Usuario (1:1 relationship) |
| **Documentos de Vehículo (drivers)** | | | |
| `vehicle_brand` | varchar(100) | nullable | Marca del vehículo |
| `vehicle_model` | varchar(100) | nullable | Modelo del vehículo |
| `license_plate` | varchar(20) | nullable | Placas |
| `circulation_card_url` | text | nullable | URL tarjeta de circulación |
| `circulation_card_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| **Documentos de Identidad** | | | |
| `ine_front_url` | text | nullable | URL INE frontal |
| `ine_front_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| `ine_back_url` | text | nullable | URL INE trasera |
| `ine_back_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| `driver_license_front_url` | text | nullable | URL licencia de conducir frontal |
| `driver_license_front_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| `driver_license_back_url` | text | nullable | URL licencia trasera |
| `driver_license_back_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| `profile_photo_url` | text | nullable | URL foto de perfil |
| `profile_photo_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| **Datos Fiscales (merchants)** | | | |
| `fiscal_name` | varchar(255) | nullable | Razón social |
| `fiscal_name_encrypted` | bytea | nullable | 🔒 Razón social cifrada (AES-256) |
| `fiscal_rfc` | varchar(13) | nullable | RFC (12-13 caracteres) |
| `fiscal_rfc_encrypted` | bytea | nullable | 🔒 RFC cifrado (AES-256) |
| `fiscal_zip_code` | varchar(5) | nullable | Código postal fiscal |
| `fiscal_regime` | enum fiscal_regime_type | nullable | Régimen fiscal |
| `fiscal_street` | varchar(255) | nullable | Calle |
| `fiscal_street_encrypted` | bytea | nullable | 🔒 Calle cifrada (AES-256) |
| `fiscal_ext_number` | varchar(20) | nullable | Número exterior |
| `fiscal_int_number` | varchar(20) | nullable | Número interior |
| `fiscal_neighborhood` | varchar(100) | nullable | Colonia |
| `fiscal_neighborhood_encrypted` | bytea | nullable | 🔒 Colonia cifrada (AES-256) |
| `fiscal_city` | varchar(100) | nullable | Ciudad |
| `fiscal_city_encrypted` | bytea | nullable | 🔒 Ciudad cifrada (AES-256) |
| `fiscal_state` | varchar(100) | nullable | Estado |
| `fiscal_certificate_url` | text | nullable | URL constancia de situación fiscal |
| `fiscal_certificate_url_encrypted` | bytea | nullable | 🔒 URL cifrada (AES-256) |
| **Metadatos** | | | |
| `reviewed` | boolean | DEFAULT false, NOT NULL | Si admin ya revisó los documentos |
| `created_at` | timestamptz | DEFAULT NOW(), NOT NULL | Fecha de creación |
| `updated_at` | timestamptz | DEFAULT NOW(), NOT NULL | Última actualización |

### Índices

- `user_documents_pkey`: PRIMARY KEY (id)
- `user_documents_user_id_key`: UNIQUE (user_id)
- `idx_user_documents_user_id`: (user_id)
- `idx_user_documents_reviewed`: (reviewed)

### Relaciones

**References:**
- `user_id` → `users(id)` CASCADE DELETE

### Enum: fiscal_regime_type

```sql
CREATE TYPE fiscal_regime_type AS ENUM (
  'persona_fisica_actividad_empresarial',
  'regimen_simplificado_confianza',
  'arrendamiento',
  'actividad_profesional',
  'persona_moral'
);
```

### Flujo de Verificación KYC

**Para Drivers:**
1. Driver sube documentos:
   - INE (frontal + trasera)
   - Licencia de conducir (frontal + trasera)
   - Tarjeta de circulación
   - Foto de perfil
2. Documentos se almacenan en R2: `https://delivery.tacoshare.app/documents/{user_id}/{filename}`
3. `reviewed = false` hasta que admin revise
4. Admin panel: `GET /api/v1/documents/{user_id}` → revisar → `PATCH /api/v1/documents/{user_id}/review`
5. Si aprobado: `users.account_status = 'active'`

**Para Merchants:**
1. Merchant sube documentos fiscales:
   - RFC
   - Constancia de situación fiscal
2. Misma lógica de revisión

### Admin Panel Query

```sql
-- Documentos pendientes de revisión
SELECT u.id, u.name, u.email, u.role, ud.created_at
FROM user_documents ud
JOIN users u ON u.id = ud.user_id
WHERE ud.reviewed = false
ORDER BY ud.created_at ASC;
```

---

## 7. notifications

**Descripción:** Notificaciones push enviadas a usuarios. Se almacenan para historial y lectura posterior. Se envían vía Firebase Cloud Messaging (FCM).

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `user_id` | uuid | FK users(id), NOT NULL | Usuario destinatario |
| `title` | varchar(255) | NOT NULL | Título de la notificación |
| `body` | text | NOT NULL | Cuerpo del mensaje |
| `data` | jsonb | nullable | Datos adicionales (order_id, deep_link, etc.) |
| `is_read` | boolean | DEFAULT false, NOT NULL | Si el usuario la leyó |
| `created_at` | timestamptz | DEFAULT NOW(), NOT NULL | Fecha de envío |
| `updated_at` | timestamptz | DEFAULT NOW(), NOT NULL | Última actualización |
| `notification_type` | varchar(50) | DEFAULT 'general', NOT NULL | Tipo de notificación |
| `read_at` | timestamptz | nullable | Fecha de lectura |

### Índices

- `notifications_pkey`: PRIMARY KEY (id)
- `idx_notifications_user_id`: (user_id)
- `idx_notifications_is_read`: (is_read)
- `idx_notifications_type`: (notification_type)
- `idx_notifications_created_at`: (created_at DESC)
- `idx_notifications_user_id_created_at`: (user_id, created_at DESC)

### Constraints

- **CHECK `notifications_notification_type_check`**: 
  ```
  notification_type IN (
    'order_created', 'order_updated', 'order_assigned', 
    'order_in_transit', 'order_delivered', 'order_cancelled',
    'payment_received', 'payment_failed',
    'driver_assigned', 'driver_nearby',
    'general', 'promotional'
  )
  ```

### Relaciones

**References:**
- `user_id` → `users(id)` CASCADE DELETE

### Triggers

- `trigger_update_notifications_updated_at`: Actualiza `updated_at` en cada UPDATE

### Row-Level Security (RLS)

**Políticas habilitadas:**
- ✅ **SELECT**: Solo el propio usuario puede ver sus notificaciones
- ✅ **INSERT**: Sistema puede crear notificaciones para cualquier usuario
- ✅ **UPDATE**: Solo el propio usuario puede actualizar sus notificaciones (marcar como leídas)
- ✅ **DELETE**: Solo el propio usuario puede eliminar sus notificaciones
- ✅ **Admin override**: Rol admin puede ver todas las notificaciones

### Tipos de Notificaciones

| Tipo | Destinatario | Descripción |
|------|-------------|-------------|
| `order_created` | Merchant | Nueva orden creada |
| `order_updated` | Merchant, Driver | Cambio de estado de orden |
| `order_assigned` | Driver | Te asignaron una orden (15s para aceptar) |
| `order_in_transit` | Customer | Conductor en camino |
| `order_delivered` | Merchant, Customer | Orden entregada |
| `order_cancelled` | Driver, Merchant | Orden cancelada |
| `payment_received` | Merchant | Pago recibido exitosamente |
| `payment_failed` | Merchant | Fallo en el pago |
| `driver_assigned` | Customer | Conductor asignado |
| `driver_nearby` | Customer | Conductor cerca (geofence) |
| `general` | Todos | Notificaciones generales |
| `promotional` | Todos | Promociones y ofertas |

### Estructura del Campo `data` (JSONB)

**Ejemplo para `order_assigned`:**
```json
{
  "order_id": "550e8400-e29b-41d4-a716-446655440000",
  "merchant_name": "Tacos Don Pepe",
  "pickup_address": "Av. Reforma 123",
  "delivery_address": "Insurgentes 456",
  "total_amount": 250.50,
  "distance_km": 3.2,
  "deep_link": "tacoshare://orders/550e8400-e29b-41d4-a716-446655440000"
}
```

### API Endpoints

```
GET /api/v1/notifications/me?page=1&limit=20&unread=true
PATCH /api/v1/notifications/{id}/read
PATCH /api/v1/notifications/mark-all-read
```

---

## 8. fcm_tokens

**Descripción:** Tokens FCM (Firebase Cloud Messaging) de dispositivos registrados para recibir notificaciones push. Un usuario puede tener múltiples dispositivos.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `user_id` | uuid | FK users(id), NOT NULL | Usuario propietario |
| `token` | text | NOT NULL | Token FCM del dispositivo |
| `token_encrypted` | bytea | nullable | 🔒 Token cifrado (AES-256) |
| `device_type` | varchar(20) | NOT NULL | Tipo: android, ios, web |
| `invalid` | boolean | DEFAULT false | 🔒 Si el token fue marcado como inválido por FCM |
| `created_at` | timestamptz | DEFAULT NOW(), NOT NULL | Fecha de registro |
| `updated_at` | timestamptz | DEFAULT NOW(), NOT NULL | Última actualización |

### Índices

- `fcm_tokens_pkey`: PRIMARY KEY (id)
- `fcm_tokens_user_id_token_key`: UNIQUE (user_id, token)
- `idx_fcm_tokens_user_id`: (user_id)

### Constraints

- **CHECK `fcm_tokens_device_type_check`**: device_type IN ('android', 'ios', 'web')

### Relaciones

**References:**
- `user_id` → `users(id)` CASCADE DELETE

### Triggers

- `trigger_update_fcm_tokens_updated_at`: Actualiza `updated_at` en cada UPDATE

### Row-Level Security (RLS)

**Políticas habilitadas:**
- ✅ **SELECT**: Solo el propio usuario puede ver sus tokens FCM
- ✅ **INSERT**: Solo el propio usuario puede registrar sus tokens
- ✅ **UPDATE**: Solo el propio usuario puede actualizar sus tokens
- ✅ **DELETE**: Solo el propio usuario puede eliminar sus tokens
- ✅ **Admin override**: Rol admin puede gestionar todos los tokens

### Flujo de Notificaciones (Seguro)

1. **Registro de Token:**
   - Usuario abre la app → Firebase SDK genera token
   - App envía token: `POST /api/v1/notifications/register-token`
   - Token se cifra con AES-256 y se almacena en `token_encrypted`
   - Se inserta o actualiza en `fcm_tokens` con `invalid = false`

2. **Envío de Notificación:**
   ```go
   // Obtener tokens válidos del usuario
   tokens := GetValidFCMTokensByUserID(userID) // WHERE invalid = false
   
   // Enviar a todos los dispositivos
   for _, token := range tokens {
       err := fcm.Send(decryptToken(token), notification)
       if err == FCMInvalidToken {
           markTokenAsInvalid(token.ID)
       }
   }
   ```

3. **Tokens Inválidos:**
   - Si FCM retorna `InvalidRegistration` o `NotRegistered`
   - Marcar `invalid = true` en lugar de eliminar (auditoría)
   - Limpieza automática: función `cleanup_invalid_fcm_tokens()` (cron diario)
   - Elimina tokens con `invalid = true` AND `updated_at < NOW() - 30 days`

### Protección de Tokens FCM

- ✅ **Encriptación**: Tokens almacenados cifrados (AES-256)
- ✅ **Soft delete**: Marcado como `invalid` en lugar de eliminación inmediata
- ✅ **Limpieza automática**: Cron job diario elimina tokens inválidos antiguos
- ✅ **RLS**: Solo el propietario puede acceder a sus tokens

### API Endpoints

```
POST /api/v1/notifications/register-token
Body: { "token": "fcm_token_string", "device_type": "android" }

DELETE /api/v1/notifications/unregister-token
Body: { "token": "fcm_token_string" }
```

---

## 9. refresh_tokens

**Descripción:** Tokens de refresco JWT para obtener nuevos access tokens sin volver a autenticarse. Implementa rotación de tokens y revocación.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | uuid | PK, NOT NULL | Identificador único |
| `user_id` | uuid | FK users(id), NOT NULL | Usuario propietario |
| `token_hash` | varchar(64) | UNIQUE, NOT NULL | SHA-256 del refresh token |
| `device_info` | text | nullable | User-Agent o identificador de dispositivo |
| `device_id` | varchar(255) | nullable | 🔒 ID único del dispositivo (device binding) |
| `ip_address` | varchar(45) | nullable | IP desde la que se emitió (IPv4/IPv6) |
| `expires_at` | timestamptz | NOT NULL | Fecha de expiración (90 días típicamente) |
| `created_at` | timestamptz | DEFAULT NOW() | Fecha de emisión |
| `last_used_at` | timestamptz | nullable | 🔒 Última vez que se usó el token (theft detection) |
| `revoked` | boolean | DEFAULT false | Si fue revocado manualmente |
| `revoked_at` | timestamptz | nullable | Fecha de revocación |
| `revoked_reason` | varchar(100) | nullable | 🔒 Razón de revocación (theft_detected, device_mismatch, etc.) |
| `deleted_at` | timestamptz | nullable | 🔒 Soft delete para compliance (NULL = activo) |

### Índices

- `refresh_tokens_pkey`: PRIMARY KEY (id)
- `refresh_tokens_token_hash_key`: UNIQUE (token_hash)
- `idx_refresh_tokens_user_id`: (user_id)
- `idx_refresh_tokens_token_hash`: (token_hash)
- `idx_refresh_tokens_expires_at`: (expires_at)
- `idx_refresh_tokens_revoked`: (revoked) WHERE revoked = false
- `idx_refresh_tokens_device_id`: 🔒 (user_id, device_id) WHERE device_id IS NOT NULL
- `idx_refresh_tokens_last_used`: 🔒 (last_used_at) WHERE last_used_at IS NOT NULL
- `idx_refresh_tokens_deleted_at`: 🔒 (deleted_at) WHERE deleted_at IS NULL

### Relaciones

**References:**
- `user_id` → `users(id)` CASCADE DELETE

### Row-Level Security (RLS)

**Políticas habilitadas:**
- ✅ **SELECT**: Solo el propio usuario puede ver sus refresh tokens
- ✅ **INSERT**: Solo el propio usuario puede crear sus refresh tokens
- ✅ **UPDATE**: Solo el propio usuario puede actualizar sus refresh tokens
- ✅ **DELETE**: Solo el propio usuario puede eliminar sus refresh tokens
- ✅ **Admin override**: Rol admin puede gestionar todos los tokens

### Seguridad de Tokens

**Nunca almacenar tokens en texto plano:**
```go
// Al crear refresh token
plainToken := generateRandomToken() // 32 bytes random (crypto/rand)
tokenHash := sha256(plainToken)
storeInDB(tokenHash) // Solo guardar el hash
returnToClient(plainToken) // Enviar texto plano al cliente

// Al validar refresh token
plainToken := fromRequest()
tokenHash := sha256(plainToken)
tokenRecord := findByHash(tokenHash)
```

### Flujo de Refresh (Fortificado)

1. **Login exitoso:**
   - Generar `access_token` (expira 15 min) + `refresh_token` (expira 90 días)
   - Guardar `SHA256(refresh_token)` en tabla con `device_id` y `ip_address`
   - `last_used_at = created_at`, `revoked = false`
   - Retornar ambos tokens al cliente

2. **Access token expiró:**
   - Cliente envía `refresh_token` + `device_id` a `POST /api/v1/auth/refresh`
   - Validar hash en DB
   - **VERIFICACIÓN 1**: `revoked = false` y `deleted_at IS NULL`
   - **VERIFICACIÓN 2**: `expires_at > NOW()`
   - **VERIFICACIÓN 3 (Device Binding)**: Si `device_id` está registrado, debe coincidir
   - **VERIFICACIÓN 4 (Theft Detection)**: Si token ya fue revocado → revocar TODOS los tokens del usuario
   - Actualizar `last_used_at = NOW()`
   - Generar nuevo `access_token` + nuevo `refresh_token` (rotation)
   - Revocar el viejo refresh token con `revoked_reason = 'rotated'`
   - Retornar nuevos tokens

3. **Logout:**
   - `DELETE /api/v1/auth/logout`
   - Marcar `revoked = true`, `revoked_at = NOW()`, `revoked_reason = 'user_logout'`

4. **Logout de todos los dispositivos:**
   - `DELETE /api/v1/auth/logout-all`
   - Revocar todos los refresh tokens del usuario con `revoked_reason = 'logout_all'`

### Limpieza Automática

```sql
-- Cron job diario: soft delete tokens expirados
UPDATE refresh_tokens
SET deleted_at = NOW()
WHERE deleted_at IS NULL
  AND (expires_at < NOW() - INTERVAL '30 days' OR (revoked = true AND revoked_at < NOW() - INTERVAL '30 days'));

-- Cron job mensual: hard delete tokens antiguos
DELETE FROM refresh_tokens
WHERE deleted_at < NOW() - INTERVAL '90 days';
```

### Seguridad Avanzada

**1. Device Binding:**
- Cada refresh token vinculado a un `device_id` único
- Si el token se usa desde otro dispositivo → revocación automática
- Previene robo de tokens cross-device

**2. Theft Detection (Detección de Robo):**
- Si se intenta usar un refresh token ya revocado → **TOKEN THEFT DETECTED**
- Acción inmediata: revocar TODOS los tokens del usuario con `revoked_reason = 'theft_detected'`
- Usuario forzado a re-login en todos los dispositivos

**3. Reuse Detection (Detección de Reuso):**
- Campo `last_used_at` registra cada uso del token
- Si un token revocado se intenta reusar → activar theft detection
- Previene ataques de replay

**4. Razones de Revocación:**
- `rotated`: Token rotado normalmente
- `user_logout`: Usuario cerró sesión
- `logout_all`: Usuario cerró sesión en todos los dispositivos
- `theft_detected`: Detección de robo de token
- `device_mismatch`: Device binding falló
- `expired`: Token expirado
- `admin_revoke`: Revocado por administrador

### Estadísticas de Seguridad

```sql
-- Tokens robados en las últimas 24 horas
SELECT COUNT(*) FROM refresh_tokens
WHERE revoked_reason = 'theft_detected'
  AND revoked_at >= NOW() - INTERVAL '24 hours';

-- Intentos de reuso de tokens revocados
SELECT user_id, COUNT(*) AS attempts
FROM refresh_tokens
WHERE revoked = true AND last_used_at > revoked_at
GROUP BY user_id
ORDER BY attempts DESC;
```

---

## 10. schema_migrations

**Descripción:** Control de versiones de migraciones de base de datos. Gestionada automáticamente por `golang-migrate`.

### Estructura

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `version` | bigint | PK, NOT NULL | Número de versión de la migración |
| `dirty` | boolean | NOT NULL | Si la migración falló a mitad de ejecución |

### Ejemplo de Registros

```
 version |  dirty
---------+---------
       1 | f
       2 | f
       3 | f
      11 | f
```

### Estado "Dirty"

**Si `dirty = true`:**
- La migración falló a mitad de ejecución
- La base de datos está en estado inconsistente
- Se debe revisar manualmente y forzar versión:

```bash
# Ver estado actual
migrate -path migrations -database $DATABASE_URL version

# Forzar versión (CUIDADO)
migrate -path migrations -database $DATABASE_URL force 11
```

### Comandos de Migración

```bash
# Aplicar todas las migraciones pendientes
make migrate-up

# Revertir última migración
make migrate-down

# Crear nueva migración
make migrate-new name=add_payments_table
```

---

## 🔗 Diagrama de Relaciones (ERD)

```
users (1) ─────────< (N) merchants
  │                      │
  │                      │
  ├─< driver_locations   ├─< orders (M) ───< (N) order_assignments ──> (M) users (drivers)
  │   (RESTRICT) 🔒      │        │
  │                      │        └─< delivery_code_audit ──> users (attempted_by)
  ├─< user_documents     │
  │   (RESTRICT) 🔒      └─────────────────────────────────────────────────────┤
  │                                                                             │
  ├─< notifications                                                             │
  │   (RLS) 🔒                                                                  │
  ├─< fcm_tokens                                                                │
  │   (RLS, encrypted) 🔒                                                       │
  └─< refresh_tokens                                                            │
      (RLS, theft detection) 🔒                                                 │
                                                                                 │
orders (1) ──> (1) cancelled_by (users)                                         │
orders (1) ──> (1) driver_id (users) ───────────────────────────────────────────┘

audit_log (append-only, partitioned) 🔒
  └── Audita: orders, refresh_tokens, users
```

**Leyenda:**
- `(1) ─────< (N)`: One-to-Many
- `(M) ───< (N)`: Many-to-Many (con tabla intermedia)
- `──>`: Foreign Key

---

## 📊 Estadísticas de la Base de Datos

### Resumen de Tablas

| Tabla | Propósito | Índices | Foreign Keys | Triggers | Seguridad |
|-------|-----------|---------|--------------|----------|-----------|
| `users` | Autenticación y roles | 9 | 0 | 0 | 🔒 OTP hash, rate limiting, soft delete |
| `merchants` | Negocios | 6 | 1 | 0 | ✅ |
| `orders` | Órdenes de delivery | 11 | 3 | 0 | 🔒 Delivery code crypto, PII encrypted, state machine |
| `driver_locations` | GPS en tiempo real | 7 | 1 (RESTRICT) | 0 | 🔒 RESTRICT delete |
| `order_assignments` | Historial de asignaciones | 7 | 2 | 0 | ✅ |
| `user_documents` | KYC | 3 | 1 (RESTRICT) | 0 | 🔒 RLS, PII encrypted, RESTRICT delete |
| `notifications` | Push notifications | 6 | 1 | 1 | 🔒 RLS |
| `fcm_tokens` | Tokens de dispositivos | 3 | 1 | 1 | 🔒 RLS, tokens encrypted, cleanup |
| `refresh_tokens` | Tokens JWT | 9 | 1 | 1 | 🔒 RLS, device binding, theft detection, soft delete |
| `schema_migrations` | Control de versiones | 1 | 0 | 0 | ✅ |

**Total:** 10 tablas, 62+ índices, 11 foreign keys, 3 triggers

**Mejoras de seguridad implementadas:**
- ✅ Row-Level Security (RLS) en 4 tablas
- ✅ Encriptación PII (AES-256) en 3 tablas
- ✅ OTP hash SHA-256 + rate limiting
- ✅ Delivery code cryptographic generation + attempt counter
- ✅ Refresh token theft detection
- ✅ Device binding para tokens
- ✅ Soft delete en 2 tablas
- ✅ RESTRICT delete en 2 tablas
- ✅ State machine enforcement en orders

**Nota:** Las tablas `audit_log` y `delivery_code_audit` fueron removidas en migración 000022.
Rate limiting de delivery codes se maneja con `delivery_code_attempts` en la tabla `orders`.
Para auditoría en producción, se recomienda logging estructurado a nivel de aplicación.

---

## 🛠️ Mantenimiento y Optimización

### Índices Geoespaciales

**Queries comunes con índices compuestos:**

```sql
-- Buscar conductores disponibles cerca del pickup
EXPLAIN ANALYZE
SELECT * FROM driver_locations
WHERE is_available = true
  AND latitude BETWEEN 19.4 AND 19.5
  AND longitude BETWEEN -99.2 AND -99.1
ORDER BY (pickup_lat - latitude)^2 + (pickup_lng - longitude)^2 ASC
LIMIT 10;

-- Usa: idx_driver_locations_available_location
```

### Vacuuming y Autovacuum

```sql
-- Configurar autovacuum agresivo para tablas de alta escritura
ALTER TABLE driver_locations SET (
  autovacuum_vacuum_scale_factor = 0.01,
  autovacuum_analyze_scale_factor = 0.005
);

ALTER TABLE order_assignments SET (
  autovacuum_vacuum_scale_factor = 0.02
);
```

### Particionamiento (Futuro)

**Si `orders` crece > 10M rows:**

```sql
-- Particionar por mes
CREATE TABLE orders_2025_01 PARTITION OF orders
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE orders_2025_02 PARTITION OF orders
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
```

### Limpieza Automática

**Scripts recomendados (cron jobs):**

```sql
-- 1. Eliminar refresh tokens expirados (diario)
DELETE FROM refresh_tokens
WHERE expires_at < NOW() - INTERVAL '30 days';

-- 2. Archivar órdenes antiguas (mensual)
INSERT INTO orders_archive
SELECT * FROM orders
WHERE created_at < NOW() - INTERVAL '6 months';

DELETE FROM orders
WHERE created_at < NOW() - INTERVAL '6 months';

-- 3. Limpiar ubicaciones obsoletas de drivers (cada hora)
DELETE FROM driver_locations
WHERE updated_at < NOW() - INTERVAL '24 hours';
```

---

## 🔐 Seguridad

### Datos Sensibles

**Campos con encriptación en reposo (AES-256 con pgcrypto):**
- `users.phone_encrypted` (teléfono cifrado)
- `orders.customer_phone_encrypted` (teléfono del cliente)
- `orders.pickup_address_encrypted` (dirección de recogida)
- `orders.delivery_address_encrypted` (dirección de entrega)
- `fcm_tokens.token_encrypted` (tokens FCM cifrados)
- `user_documents.fiscal_rfc_encrypted` (RFC fiscal)
- `user_documents.*_url_encrypted` (URLs de documentos sensibles)

**Campos con hash criptográfico:**
- `users.password_hash` (bcrypt con salt automático)
- `users.otp_hash` (SHA-256 + pepper del servidor)
- `refresh_tokens.token_hash` (SHA-256)

**Campos con expiración temporal:**
- `users.otp_hash` (expiración de 10 min, limpieza automática)
- `users.otp_locked_until` (lockout de 15 min tras 3 intentos fallidos)
- `refresh_tokens.expires_at` (90 días)

**Funciones de encriptación disponibles:**
```sql
-- Cifrar
SELECT encrypt_text('dato sensible') AS encrypted;

-- Descifrar
SELECT decrypt_text(columna_encrypted) FROM tabla;
```

### Auditoría

**Campos de auditoría presentes:**
- `created_at`: Todas las tablas
- `updated_at`: Todas las tablas (excepto `schema_migrations`)
- `revoked_at`: `refresh_tokens`
- `responded_at`: `order_assignments`
- `last_used_at`: `refresh_tokens` (theft detection)
- `deleted_at`: `users`, `refresh_tokens` (soft delete)
- Timestamps de estado: `assigned_at`, `accepted_at`, `delivered_at`, etc.

**Rate limiting de delivery codes:**
- `orders.delivery_code_attempts`: Contador de intentos fallidos (máx. 3)
- Rate limiting implementado en capa de aplicación
- Logs estructurados para detección de patrones de ataque

**Nota sobre auditoría:**
Las tablas `audit_log` y `delivery_code_audit` fueron removidas en migración 000022 por complejidad operativa y crecimiento sin límite. Para producción se recomienda:
- Logging estructurado a nivel de aplicación (CloudWatch, Loki, etc.)
- WAL archiving de PostgreSQL para point-in-time recovery
- Auditoría específica solo para eventos críticos de compliance

### Constraints de Integridad

**Validaciones a nivel de BD:**
- Email único: `users.email` UNIQUE
- Teléfono único: `users.phone` UNIQUE
- Roles válidos: `users.role` CHECK
- Estados de orden válidos: `orders.status` CHECK
- Monto positivo: `orders.total_amount` > 0
- Código de delivery de 4 dígitos: `orders.delivery_code` REGEX
- Rating 0-5: `merchants.rating` CHECK

---

## 📈 Queries Comunes de Análisis

### Métricas de Negocio

```sql
-- Órdenes completadas por merchant (últimos 30 días)
SELECT m.business_name, COUNT(*) AS total_orders, SUM(o.total_amount) AS revenue
FROM orders o
JOIN merchants m ON m.id = o.merchant_id
WHERE o.status = 'delivered'
  AND o.delivered_at >= NOW() - INTERVAL '30 days'
GROUP BY m.business_name
ORDER BY revenue DESC;

-- Tasa de aceptación por conductor
SELECT u.name, 
  COUNT(*) FILTER (WHERE oa.status = 'accepted') AS accepted,
  COUNT(*) FILTER (WHERE oa.status IN ('rejected', 'timeout')) AS rejected,
  ROUND(COUNT(*) FILTER (WHERE oa.status = 'accepted') * 100.0 / COUNT(*), 2) AS acceptance_rate
FROM order_assignments oa
JOIN users u ON u.id = oa.driver_id
GROUP BY u.name
HAVING COUNT(*) >= 10
ORDER BY acceptance_rate DESC;

-- Tiempo promedio de entrega por ciudad
SELECT m.city,
  AVG(EXTRACT(EPOCH FROM (o.delivered_at - o.created_at)) / 60) AS avg_delivery_minutes
FROM orders o
JOIN merchants m ON m.id = o.merchant_id
WHERE o.status = 'delivered'
  AND o.delivered_at >= NOW() - INTERVAL '7 days'
GROUP BY m.city
ORDER BY avg_delivery_minutes ASC;

-- Órdenes sin conductor disponible (por día)
SELECT DATE(created_at) AS day, COUNT(*) AS no_driver_orders
FROM orders
WHERE status = 'no_driver_available'
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY day DESC;
```

---

## 🚀 Migraciones Pendientes (Roadmap)

### Próximas Features (Tablas a Crear)

1. **`payments`** - Pagos con Stripe Connect
   - `order_id`, `stripe_payment_intent_id`, `amount`, `status`, `metadata`

2. **`ratings`** - Calificaciones de drivers/merchants
   - `order_id`, `rater_id`, `rated_id`, `rating`, `comment`

3. **`driver_earnings`** - Ganancias de conductores
   - `driver_id`, `order_id`, `base_fee`, `distance_bonus`, `total`, `paid_out`

4. **`promotions`** - Códigos promocionales
   - `code`, `discount_type`, `discount_value`, `expires_at`, `max_uses`

5. **`order_tracking`** - Tracking de ubicación del driver durante entrega
   - `order_id`, `driver_id`, `latitude`, `longitude`, `timestamp`

---

## 📚 Recursos Adicionales

- **Migraciones:** `/migrations/*.sql`
- **Esquemas Go:** `/internal/*/models/*.go`
- **OpenAPI Docs:** `http://localhost:8080/docs` (Scalar UI)
- **Swagger JSON:** `http://localhost:8080/swagger/doc.json`

---

**Última actualización:** 2025-01-27  
**PostgreSQL Version:** 18 Alpine  
**Migraciones aplicadas:** 22 (última: remove_unused_audit_tables)  
**Estado:** ✅ MVP Ready
