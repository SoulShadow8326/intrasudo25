#!/bin/bash

echo "Stopping all intrasudo25 services..."

pkill -f "intrasudo25" || true
pkill -f "loadbalancer" || true
pkill -f "bot.py" || true

echo "All services stopped."
