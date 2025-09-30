#!/bin/bash
#

echo "Waiting for database to be ready ..."

until pg_isready -h db -U $GOBID_DATABASE_USER -d $GOBID_DATABASE_NAME; do
	sleep 1
done


echo "Creating database if does not exist ..."

PGPASSWORD=$GOBID_DATABASE_PASSWORD psql -h db -U $GOBID_DATABASE_USER -c "CREATE DATABASE $GOBID_DATABASE_NAME;" || true

echo "Applying migrations"

cd /app/internal/store/pgstore/migrations
tern migrate

echo "Starting app"
cd /app

exec ./api
