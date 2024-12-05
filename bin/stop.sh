#!/bin/bash

# Get the directory where the script is located
APP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_NAME="echogy"
PID_FILE="$APP_DIR/$APP_NAME.pid"

# Check if PID file exists
if [ ! -f "$PID_FILE" ]; then
    echo "PID file not found. $APP_NAME may not be running."
    exit 1
fi

# Read PID from file
PID=$(cat "$PID_FILE")

# Check if process is running
if ! ps -p $PID > /dev/null 2>&1; then
    echo "Process not found. Removing stale PID file."
    rm "$PID_FILE"
    exit 1
fi

# Send SIGTERM signal to the process
echo "Stopping $APP_NAME (PID: $PID)..."
kill $PID

# Wait for the process to stop
TIMEOUT=30
while ps -p $PID > /dev/null 2>&1; do
    if [ "$TIMEOUT" -le 0 ]; then
        echo "Timeout waiting for $APP_NAME to stop. Forcing shutdown..."
        kill -9 $PID
        break
    fi
    sleep 1
    ((TIMEOUT--))
done

# Check if PID file was removed by the application
if [ -f "$PID_FILE" ]; then
    rm "$PID_FILE"
fi

echo "$APP_NAME stopped successfully"
