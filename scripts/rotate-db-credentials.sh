#!/bin/bash

# Script de rotación de credenciales de base de datos
# IMPORTANTE: Ejecutar en ambiente seguro, nunca commitear credenciales

set -e

echo "🔄 Rotación de credenciales de base de datos TacoShare"
echo "=================================================="
echo ""

# Validar que estamos en el directorio correcto
if [ ! -f ".env" ]; then
    echo "❌ Error: archivo .env no encontrado"
    echo "   Ejecutar desde el directorio raíz del proyecto"
    exit 1
fi

echo "⚠️  ADVERTENCIA: Este script rotará las credenciales de la base de datos"
echo "   Asegúrate de tener un backup antes de continuar"
echo ""
read -p "¿Continuar? (y/N): " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "❌ Rotación cancelada"
    exit 0
fi

# Generar nueva contraseña segura (32 caracteres alfanuméricos + símbolos)
NEW_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)

echo ""
echo "🔑 Nueva contraseña generada (cópiala para Secret Manager):"
echo "   $NEW_PASSWORD"
echo ""

# Leer credenciales actuales
source .env

echo "📝 Pasos a seguir:"
echo ""
echo "1. Crear nuevo usuario de aplicación (no superusuario):"
echo "   psql -U postgres -h $DB_HOST -c \"CREATE USER tacoshare_app WITH PASSWORD '$NEW_PASSWORD';\""
echo ""
echo "2. Otorgar permisos al nuevo usuario:"
echo "   psql -U postgres -h $DB_HOST -d $DB_NAME -c \"GRANT CONNECT ON DATABASE $DB_NAME TO tacoshare_app;\""
echo "   psql -U postgres -h $DB_HOST -d $DB_NAME -c \"GRANT USAGE ON SCHEMA public TO tacoshare_app;\""
echo "   psql -U postgres -h $DB_HOST -d $DB_NAME -c \"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO tacoshare_app;\""
echo "   psql -U postgres -h $DB_HOST -d $DB_NAME -c \"GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO tacoshare_app;\""
echo ""
echo "3. Actualizar .env (local):"
echo "   DB_USER=tacoshare_app"
echo "   DB_PASSWORD=$NEW_PASSWORD"
echo "   DATABASE_URL=postgres://tacoshare_app:$NEW_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
echo ""
echo "4. Actualizar Secret Manager (producción):"
echo "   - GCP: gcloud secrets versions add DATABASE_URL --data-file=-"
echo "   - AWS: aws secretsmanager update-secret --secret-id DATABASE_URL --secret-string 'postgres://tacoshare_app:$NEW_PASSWORD@...'"
echo "   - Heroku: heroku config:set DATABASE_URL='postgres://tacoshare_app:$NEW_PASSWORD@...'"
echo ""
echo "5. Revocar acceso del usuario antiguo (después de validar):"
echo "   psql -U postgres -h $DB_HOST -c \"REVOKE ALL PRIVILEGES ON DATABASE $DB_NAME FROM postgres;\""
echo "   psql -U postgres -h $DB_HOST -c \"ALTER USER postgres WITH NOLOGIN;\""
echo ""
echo "⚠️  NO ejecutar paso 5 hasta validar que la app funciona con el nuevo usuario"
echo ""
echo "✅ Guarda esta información de forma segura"
