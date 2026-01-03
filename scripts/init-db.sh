#!/bin/bash

# Database initialization helper script
# Usage: ./scripts/init-db.sh <connection_string>
# Example: ./scripts/init-db.sh "postgresql://user:password@host:5432/dbname"

if [ -z "$1" ]; then
    echo "Error: Database connection string required"
    echo "Usage: $0 <connection_string>"
    echo "Example: $0 'postgresql://user:password@host:5432/dbname'"
    exit 1
fi

CONNECTION_STRING=$1

echo "Initializing database..."
echo "Connection: ${CONNECTION_STRING%%@*}@***" # Hide credentials in output

# Run init.sql using psql
psql "$CONNECTION_STRING" -f init.sql

if [ $? -eq 0 ]; then
    echo "✓ Database initialized successfully!"
    echo ""
    echo "System accounts created:"
    echo "  1001 - Referral Funding Pool (\$10,000.00)"
    echo "  1002 - Institution Main Account"
    echo "  1003 - Institution Disbursement Account"
    echo "  1004 - PSP Account"
else
    echo "✗ Database initialization failed!"
    exit 1
fi
