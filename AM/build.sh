#!/bin/bash

set -e  # Exit immediately if any command fails

# Define paths
FAASD_DIR="$HOME/go/src/github.com/openfaas/faasd"
BIN_PATH="/usr/local/bin/faasd"

echo "🚀 Starting faasd build process..."

# Navigate to faasd source directory
cd "$FAASD_DIR" || { echo "❌ Error: Directory $FAASD_DIR not found!"; exit 1; }

echo "📦 Building faasd..."
go build -o faasd

# Stop faasd services before replacing the binary
echo "🛑 Stopping faasd services..."
sudo systemctl stop faasd || echo "⚠️ faasd service was not running."
sudo systemctl stop faasd-provider || echo "⚠️ faasd-provider service was not running."

# Move the new binary into place
echo "📁 Copying new faasd binary..."
sudo cp faasd "$BIN_PATH"

# Restart faasd services
echo "🔄 Restarting faasd services..."
sudo systemctl start faasd
sudo systemctl start faasd-provider

# Verify the update
echo "✅ faasd version:"
faasd version

echo "🎉 Build and deployment completed successfully!"
