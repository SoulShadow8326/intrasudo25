#!/bin/bash

echo "Building intrasudo25..."

mkdir -p ./dist

go build -o ./dist/intrasudo25 .

mkdir -p ./dist/data
if [ -f "./data/data.db" ]; then
    cp ./data/data.db ./dist/data/
fi

cp -r ./frontend ./dist/
cp bot.py ./dist/
cp requirements.txt ./dist/
if [ -f ".env" ]; then
    cp .env ./dist/
fi

echo "Installing Python dependencies..."
cd ./dist
pip install -r requirements.txt
cd ..

echo "Build complete!"
echo "To run main server: cd dist && ./intrasudo25"
echo "To run Discord bot: cd dist && python bot.py"
echo ""
echo "deployed with:"
echo "- ./intrasudo25 (Go executable)"
echo "- ./bot.py (Python Discord bot)"
echo "- ./data/ (database directory)"
echo "- ./frontend/ (static files)"
echo "- requirements.txt (Python dependencies)"
