#!/bin/bash

set -e

echo "Building intrasudo25 project..."

cd "$(dirname "$0")"

echo "Building main application..."
go build -o intrasudo25 .

echo "Building load balancer..."
cd loadbalancer
go build -o loadbalancer main.go config.go
cd ..

echo "Starting load balancer..."
cd loadbalancer
./loadbalancer &
LB_PID=$!
cd ..

sleep 2

echo "Starting main application..."
./intrasudo25 &
MAIN_PID=$!

sleep 2

echo "Starting Discord bot..."
python3 bot.py &
BOT_PID=$!

echo "All services started:"
echo "Load Balancer PID: $LB_PID"
echo "Main App PID: $MAIN_PID"
echo "Discord Bot PID: $BOT_PID"

echo "running"

cleanup() {
    echo "Stopping all services..."
    kill $BOT_PID 2>/dev/null || true
    kill $MAIN_PID 2>/dev/null || true
    kill $LB_PID 2>/dev/null || true
    echo "All services stopped."
    exit 0
}

trap cleanup SIGINT SIGTERM

wait
