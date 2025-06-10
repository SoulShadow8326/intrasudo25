#!/bin/bash

set -e

echo "Cleaning build artifacts..."

cd "$(dirname "$0")"

rm -f intrasudo25
rm -f loadbalancer/loadbalancer

echo "Installing Python dependencies..."
pip3 install -r requirements.txt

echo "Setting up database..."
mkdir -p ./data
if [ ! -f "./data/data.db" ]; then
    touch ./data/data.db
fi

echo "Initializing Go modules..."
go mod tidy

echo "Setup complete. Run ./run.sh to start all services."
