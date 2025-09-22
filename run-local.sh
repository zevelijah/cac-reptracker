#!/bin/bash

# A more robust script to run the local development environment.
#
# It performs the following actions:
# 1. Sets environment variables (and suggests using a .env.local file for secrets).
# 2. Builds and runs the Go backend server in the background.
# 3. Launches the specified Flutter emulator.
# 4. Sets up a trap to automatically shut down the backend server on exit (Ctrl+C).

# Exit immediately if a command fails.
set -e

# --- Configuration ---
PROJECT_ROOT=$(pwd)
GO_SERVER_DIR="$PROJECT_ROOT/golang/server"
FLUTTER_EMULATOR_NAME="Medium_Phone_API_36.0"
SERVER_BINARY_NAME="server"

# --- Environment ---
# For better security, avoid hardcoding secrets.
# Create a '.env.local' file in the project root for them.
# E.g.:
#   LOC_API_KEY="your_api_key_here"
# Make sure to add .env.local to your .gitignore file!
if [ -f "$PROJECT_ROOT/.env.local" ]; then
  echo "Sourcing environment variables from .env.local file..."
  set -a # Automatically export all variables from the source file
  source "$PROJECT_ROOT/.env.local"
  set +a # Stop automatically exporting
fi

# --- Cleanup Logic ---
# This function is called when the script exits (e.g., on Ctrl+C).
cleanup() {
  echo "" # Newline for cleaner output
  if [ -n "$GO_SERVER_PID" ]; then
    echo "Shutting down Go backend server (PID: $GO_SERVER_PID)..."
    # Kill the background server process.
    kill "$GO_SERVER_PID" 2>/dev/null
    echo "Server shut down."
  fi
  # Clean up the compiled binary
  rm -f "$GO_SERVER_DIR/$SERVER_BINARY_NAME"
}

# Trap the EXIT signal to run the cleanup function automatically.
trap cleanup EXIT

# --- Main Script ---

echo "Building Go backend server..."
go -C "$GO_SERVER_DIR" build -o "$SERVER_BINARY_NAME" .

echo "Starting Go backend server in '$MODE' mode..."
"$GO_SERVER_DIR/$SERVER_BINARY_NAME" &
GO_SERVER_PID=$! # Capture the Process ID (PID) of the background server.
echo "Go server started with PID: $GO_SERVER_PID"
sleep 2 # Give the server a moment to start.

echo "Launching Flutter emulator '$FLUTTER_EMULATOR_NAME'..."
flutter emulators --launch "$FLUTTER_EMULATOR_NAME" --cold

echo -e "\nDevelopment environment is running."
echo "Backend server is active in the background."
echo "Press Ctrl+C to stop the server and exit."

# Wait for the user to interrupt the script. This keeps the script alive
# so the trap can function correctly when you press Ctrl+C.
wait $GO_SERVER_PID