#!/bin/bash

# Get the directory where the script is located
APP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APP_NAME="echogy"
PID_FILE="$APP_DIR/$APP_NAME.pid"
CONFIG_FILE="$APP_DIR/config.json"

# Check if the application is already running
if [ -f "$PID_FILE" ]; then
    if ps -p $(cat "$PID_FILE") > /dev/null 2>&1; then
        echo "Error: $APP_NAME is already running with PID $(cat "$PID_FILE")"
        exit 1
    else
        # PID file exists but process is not running, remove stale PID file
        rm "$PID_FILE"
    fi
fi

# Start the application
cd "$APP_DIR"
echo "Starting $APP_NAME..."
"$APP_DIR/$APP_NAME" -c "$CONFIG_FILE" -pid "$PID_FILE" > /dev/null 2>&1 &

# Wait a moment and check if the process is running
sleep 1
if [ -f "$PID_FILE" ] && ps -p $(cat "$PID_FILE") > /dev/null 2>&1; then
    echo "$APP_NAME started successfully with PID $(cat "$PID_FILE")"
else
    echo "Error: Failed to start $APP_NAME"
    exit 1
fi
