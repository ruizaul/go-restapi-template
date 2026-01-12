#!/bin/sh
# ============================================================================
# UNIVERSAL ENTRYPOINT - Handles server mode AND migrations
# ============================================================================
# This script allows the same Docker image to run as:
# 1. API server (default): Runs /app/server
# 2. Migration job: Runs migrations when RUN_MIGRATIONS=true
#
# Usage:
#   Server mode:    docker run image
#   Migration mode: docker run -e RUN_MIGRATIONS=true image
#   Status check:   docker run -e RUN_MIGRATIONS=true -e MIGRATION_DIRECTION=status image
# ============================================================================

set -e
set -o pipefail

# ============================================================================
# CHECK MODE: Server or Migrations
# ============================================================================
if [ "${RUN_MIGRATIONS}" != "true" ]; then
    echo "🚀 Server mode - starting API server..."
    exec /app/server
fi

# ============================================================================
# MIGRATION MODE - Everything below runs only for migrations
# ============================================================================
echo "🔄 Migration mode detected - running database migrations..."

# ============================================================================
# CONFIGURATION
# ============================================================================
readonly MIGRATIONS_DIR="/app/migrations"
readonly MAX_RETRIES=3
readonly RETRY_DELAY=5
readonly TIMEOUT=300

# ============================================================================
# LOGGING FUNCTIONS
# ============================================================================
log_info() {
    echo "ℹ️  [$(date +'%Y-%m-%d %H:%M:%S')] INFO: $*"
}

log_success() {
    echo "✅ [$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS: $*"
}

log_error() {
    echo "❌ [$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}

log_warning() {
    echo "⚠️  [$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*" >&2
}

# ============================================================================
# VALIDATION
# ============================================================================
validate_environment() {
    log_info "Validando configuración del entorno..."

    if [ -z "$DATABASE_URL" ]; then
        log_error "DATABASE_URL no está configurada"
        log_error "Ejemplo: DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require"
        exit 1
    fi

    if [ ! -d "$MIGRATIONS_DIR" ]; then
        log_error "Directorio de migraciones no encontrado: $MIGRATIONS_DIR"
        exit 1
    fi

    local migration_count
    migration_count=$(find "$MIGRATIONS_DIR" -name "*.up.sql" 2>/dev/null | wc -l | tr -d ' ')
    log_info "Archivos de migración .up.sql encontrados: $migration_count"

    if [ "$migration_count" -eq 0 ]; then
        log_warning "No se encontraron archivos .up.sql en $MIGRATIONS_DIR"
    fi
}

# ============================================================================
# DATABASE CONNECTION TEST
# ============================================================================
test_database_connection() {
    log_info "Probando conexión a la base de datos..."

    local db_info
    db_info=$(echo "$DATABASE_URL" | sed 's/:\/\/[^:]*:[^@]*@/:\/\/***:***@/')
    log_info "Conectando a: $db_info"

    local attempt=1
    local max_connection_attempts=5

    while [ $attempt -le $max_connection_attempts ]; do
        local error_output
        error_output=$(timeout 10 migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1)
        local exit_code=$?
        
        # Exit code 0 = success
        # "error: no migration" = database connected but no migrations table yet (this is OK!)
        if [ $exit_code -eq 0 ] || echo "$error_output" | grep -q "error: no migration"; then
            log_success "Conexión a la base de datos exitosa"
            if echo "$error_output" | grep -q "error: no migration"; then
                log_info "Base de datos nueva detectada - se creará tabla schema_migrations"
            fi
            return 0
        else
            log_warning "Intento de conexión $attempt/$max_connection_attempts falló"
            log_error "Error de migrate: $error_output"
            if [ $attempt -lt $max_connection_attempts ]; then
                log_info "Reintentando en ${RETRY_DELAY}s..."
                sleep $RETRY_DELAY
            fi
            attempt=$((attempt + 1))
        fi
    done

    log_error "No se pudo conectar a la base de datos después de $max_connection_attempts intentos"
    return 1
}

# ============================================================================
# SMART MIGRATION DETECTION
# ============================================================================
check_pending_migrations() {
    log_info "Verificando migraciones pendientes..."

    local current_version
    current_version=$(migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1 | grep -oE '[0-9]+' | head -1 || echo "0")

    if [ -z "$current_version" ] || [ "$current_version" = "no" ]; then
        current_version="0"
    fi

    log_info "Versión actual de la base de datos: $current_version"

    local total_migrations
    total_migrations=$(find "$MIGRATIONS_DIR" -name "*.up.sql" 2>/dev/null | wc -l | tr -d ' ')

    log_info "Total de migraciones disponibles: $total_migrations"

    local pending=$((total_migrations - current_version))

    if [ "$pending" -le 0 ]; then
        log_success "✨ No hay migraciones pendientes - base de datos actualizada"
        return 1
    else
        log_info "📋 Migraciones pendientes: $pending"

        log_info "Archivos a aplicar:"
        find "$MIGRATIONS_DIR" -name "*.up.sql" 2>/dev/null | sort | tail -n "$pending" | while read -r file; do
            log_info "  - $(basename "$file")"
        done

        return 0
    fi
}

# ============================================================================
# MIGRATION EXECUTION
# ============================================================================
run_migration() {
    local direction="$1"

    log_info "Iniciando migración: $direction"

    for attempt in $(seq 1 $MAX_RETRIES); do
        log_info "Intento $attempt de $MAX_RETRIES..."

        local cmd="migrate -path $MIGRATIONS_DIR -database $DATABASE_URL"

        if [ "$direction" = "up" ]; then
            cmd="$cmd up"
        elif [ "$direction" = "down" ]; then
            cmd="$cmd down 1"
        elif [ "$direction" = "version" ] || [ "$direction" = "status" ]; then
            cmd="$cmd version"
        else
            log_error "Dirección de migración inválida: $direction"
            exit 1
        fi

        if timeout "$TIMEOUT" $cmd; then
            log_success "Migración completada exitosamente"

            local current_version
            current_version=$(migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1 || echo "unknown")
            log_info "Versión actual de la base de datos: $current_version"

            return 0
        else
            local exit_code=$?

            if [ $exit_code -eq 124 ]; then
                log_warning "Migración timeout después de ${TIMEOUT}s (intento $attempt/$MAX_RETRIES)"
            else
                log_warning "Migración falló con código de salida $exit_code (intento $attempt/$MAX_RETRIES)"
            fi

            if [ $attempt -lt $MAX_RETRIES ]; then
                log_info "Reintentando en ${RETRY_DELAY}s..."
                sleep $RETRY_DELAY
            fi
        fi
    done

    log_error "Migración falló después de $MAX_RETRIES intentos"
    return 2
}

# ============================================================================
# SHOW MIGRATION STATUS
# ============================================================================
show_status() {
    log_info "=========================================="
    log_info "Estado de Migraciones"
    log_info "=========================================="

    local current_version
    current_version=$(migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1 | grep -oE '[0-9]+' | head -1 || echo "0")

    if [ -z "$current_version" ] || [ "$current_version" = "no" ]; then
        current_version="0"
    fi

    log_info "Versión actual: $current_version"

    log_info "Migraciones disponibles:"

    find "$MIGRATIONS_DIR" -name "*.up.sql" 2>/dev/null | sort | while read -r file; do
        local filename
        filename=$(basename "$file")
        local file_version
        file_version=$(echo "$filename" | grep -oE '^[0-9]+' || echo "0")

        local status="❌ PENDIENTE"
        if [ "$file_version" -le "$current_version" ] 2>/dev/null; then
            status="✅ APLICADA"
        fi

        echo "  $status - $filename"
    done

    log_info "=========================================="
}

# ============================================================================
# MAIN MIGRATION EXECUTION
# ============================================================================
main_migration() {
    local action="${MIGRATION_DIRECTION:-up}"

    log_info "=========================================="
    log_info "TacoShare Delivery API - Smart Migrations"
    log_info "=========================================="

    validate_environment

    if ! test_database_connection; then
        log_error "Cancelando migración debido a falla de conexión"
        exit 1
    fi

    case "$action" in
        up)
            if check_pending_migrations; then
                log_info "🚀 Aplicando migraciones pendientes..."
                run_migration "up"
                exit $?
            else
                log_success "🎉 Base de datos ya está actualizada - nada que hacer"
                exit 0
            fi
            ;;
        down)
            run_migration "down"
            exit $?
            ;;
        version|status)
            show_status
            exit 0
            ;;
        *)
            log_error "Acción desconocida: $action"
            echo ""
            echo "Uso: MIGRATION_DIRECTION={up|down|status}"
            echo ""
            echo "Comandos:"
            echo "  up      Aplicar SOLO migraciones pendientes (inteligente)"
            echo "  down    Revertir última migración"
            echo "  status  Mostrar versión actual y estado de migraciones"
            exit 1
            ;;
    esac
}

# ============================================================================
# ENTRY POINT - Run migrations
# ============================================================================
main_migration
