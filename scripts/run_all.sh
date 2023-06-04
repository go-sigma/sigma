#!/bin/bash

SCRIPT_DIR=$(dirname "$0")

echo "Starting minio..."
"$SCRIPT_DIR/run_minio.sh"

echo "Starting redis..."
"$SCRIPT_DIR/run_redis.sh"

echo "Starting mysql..."
"$SCRIPT_DIR/run_mysql.sh"

echo "Starting postgresql..."
"$SCRIPT_DIR/run_postgres.sh"
