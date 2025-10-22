#!/bin/bash
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

# Custos Signer Service management script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SIGNER_HOME="$(dirname "$SCRIPT_DIR")"
PID_FILE="$SIGNER_HOME/signer.pid"
DAEMON_SCRIPT="$SCRIPT_DIR/signer-daemon.sh"

# Function to check if service is running
is_running() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0
        else
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# Function to start the service
start() {
    if is_running; then
        echo "Custos Signer Service is already running (PID: $(cat "$PID_FILE"))"
        return 0
    fi

    echo "Starting Custos Signer Service..."
    nohup "$DAEMON_SCRIPT" > "$SIGNER_HOME/logs/signer.out" 2>&1 &
    local pid=$!
    echo $pid > "$PID_FILE"
    
    # Wait a moment and check if it's still running
    sleep 2
    if is_running; then
        echo "Custos Signer Service started successfully (PID: $pid)"
        echo "Logs: $SIGNER_HOME/logs/signer.out"
        echo "HTTP endpoint: http://localhost:8084/actuator/health"
        echo "gRPC endpoint: localhost:9095"
    else
        echo "Failed to start Custos Signer Service"
        rm -f "$PID_FILE"
        return 1
    fi
}

# Function to stop the service
stop() {
    if ! is_running; then
        echo "Custos Signer Service is not running"
        return 0
    fi

    local pid=$(cat "$PID_FILE")
    echo "Stopping Custos Signer Service (PID: $pid)..."
    
    kill "$pid"
    
    # Wait for graceful shutdown
    local count=0
    while [ $count -lt 30 ] && is_running; do
        sleep 1
        count=$((count + 1))
    done
    
    if is_running; then
        echo "Force killing Custos Signer Service..."
        kill -9 "$pid"
        sleep 1
    fi
    
    if ! is_running; then
        echo "Custos Signer Service stopped"
        rm -f "$PID_FILE"
    else
        echo "Failed to stop Custos Signer Service"
        return 1
    fi
}

# Function to restart the service
restart() {
    stop
    sleep 2
    start
}

# Function to show status
status() {
    if is_running; then
        local pid=$(cat "$PID_FILE")
        echo "Custos Signer Service is running (PID: $pid)"
        echo "HTTP endpoint: http://localhost:8084/actuator/health"
        echo "gRPC endpoint: localhost:9095"
    else
        echo "Custos Signer Service is not running"
    fi
}

# Function to show logs
logs() {
    if [ -f "$SIGNER_HOME/logs/signer.out" ]; then
        tail -f "$SIGNER_HOME/logs/signer.out"
    else
        echo "No log file found at $SIGNER_HOME/logs/signer.out"
    fi
}

# Main script logic
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the Custos Signer Service"
        echo "  stop    - Stop the Custos Signer Service"
        echo "  restart - Restart the Custos Signer Service"
        echo "  status  - Show the status of the service"
        echo "  logs    - Show and follow the service logs"
        exit 1
        ;;
esac

exit $?
