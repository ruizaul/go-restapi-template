#!/bin/bash
# ============================================================================
# Fix Dirty Migration State - Cloud Shell Script
# ============================================================================
# This script fixes "Dirty database version" errors by connecting directly
# to Cloud SQL and resetting the migration state.
#
# Usage:
#   1. Upload this script to Google Cloud Shell
#   2. Run: chmod +x scripts/db_migration_fix.sh
#   3. Run: ./scripts/db_migration_fix.sh
# ============================================================================

set -e

PROJECT_ID="delivery-93190"
INSTANCE_NAME="delivery-93190"
DATABASE_NAME="taco_delivery"
DB_USER="postgres"

echo "=============================================="
echo "Fix Dirty Migration State"
echo "=============================================="
echo ""
echo "⚠️  This will reset the 'dirty' flag in schema_migrations"
echo "📊 Current issue: Dirty database version 2"
echo ""

# Check if we're in Cloud Shell
if [ -z "$CLOUD_SHELL" ]; then
    echo "❌ This script should be run in Google Cloud Shell"
    echo ""
    echo "📋 Manual steps:"
    echo "1. Go to: https://shell.cloud.google.com"
    echo "2. Upload scripts/db_migration_fix.sh"
    echo "3. Run: chmod +x scripts/db_migration_fix.sh"
    echo "4. Run: ./scripts/db_migration_fix.sh"
    exit 1
fi

echo "✅ Cloud Shell detected"
echo ""

# Set the project
echo "🔧 Setting project to $PROJECT_ID..."
gcloud config set project $PROJECT_ID

echo ""
echo "🔌 Connecting to Cloud SQL instance: $INSTANCE_NAME"
echo "📝 Database: $DATABASE_NAME"
echo ""
echo "The following SQL commands will be executed:"
echo "  1. SELECT * FROM schema_migrations; (check current state)"
echo "  2. UPDATE schema_migrations SET dirty = false WHERE version = 2; (fix dirty state)"
echo "  3. SELECT * FROM schema_migrations; (verify fix)"
echo ""

read -p "Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cancelled"
    exit 1
fi

echo ""
echo "🚀 Connecting to database..."
echo ""

# Execute SQL commands via gcloud
gcloud sql connect $INSTANCE_NAME \
    --user=$DB_USER \
    --database=$DATABASE_NAME \
    --project=$PROJECT_ID \
    <<'EOSQL'

-- Show current state
\echo '📊 Current migration state:'
SELECT version, dirty FROM schema_migrations;

-- Fix dirty flag
\echo ''
\echo '🔧 Fixing dirty flag for version 2...'
UPDATE schema_migrations SET dirty = false WHERE version = 2;

-- Verify
\echo ''
\echo '✅ New migration state:'
SELECT version, dirty FROM schema_migrations;

\echo ''
\echo '✅ Done! The dirty flag has been reset.'
\echo 'You can now run: make migrate-prod-run'

EOSQL

echo ""
echo "=============================================="
echo "✅ Migration state fixed!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "1. Go back to your local terminal"
echo "2. Run: make migrate-prod-run"
echo ""
