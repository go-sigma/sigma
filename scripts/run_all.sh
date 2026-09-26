#!/bin/bash

SCRIPT_DIR=$(dirname "$0")

echo "Starting rs3..."
"$SCRIPT_DIR/run_rs3.sh"

echo "Starting redis..."
"$SCRIPT_DIR/run_redis.sh"

echo "Starting mysql..."
"$SCRIPT_DIR/run_mysql.sh"

echo "Starting postgresql..."
"$SCRIPT_DIR/run_postgres.sh"
