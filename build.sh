#!/bin/bash

echo "Building intrasudo25..."

mkdir -p ./dist

go build -o ./dist/intrasudo25 .

mkdir -p ./dist/data
if [ -f "./data/logins.db" ]; then
    cp ./data/logins.db ./dist/data/
fi

cp -r ./frontend ./dist/

echo "Build complete! Executable is in ./dist/intrasudo25"
echo "To run: cd dist && ./intrasudo25"
echo ""
echo "The application is now stateless and can be deployed anywhere with:"
echo "- ./intrasudo25 (executable)"
echo "- ./data/ (database directory)"
echo "- ./frontend/ (static files)"
