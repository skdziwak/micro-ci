#!/bin/bash

# Exit on any error
set -e

# Define the program name and the installation directory
PROGRAM_NAME="micro-ci"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="${PROGRAM_NAME}.service"
SYSTEMD_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/${PROGRAM_NAME}"
USER_NAME="micro-ci"
REBUILD=false

if [ "$1" == "--rebuild" ]; then
  REBUILD=true
fi
if [ ! -f "$PROGRAM_NAME" ]; then
  REBUILD=true
fi
if [ "$REBUILD" = true ]; then
  echo "Building the binary..."
  go build -o ${PROGRAM_NAME} main.go
fi

if systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "Service is running and will be restarted."
    sudo systemctl stop ${SERVICE_NAME}
fi

echo "Installing the binary..."
sudo cp ${PROGRAM_NAME} ${INSTALL_DIR}/${PROGRAM_NAME}

if ! id "$USER_NAME" &>/dev/null; then
    sudo useradd --system --home ${CONFIG_DIR} --shell /sbin/nologin ${USER_NAME}
fi

if [ -d "${CONFIG_DIR}" ]; then
    echo "Configuration directory already exists."
else
    echo "Creating configuration directory..."
    sudo mkdir -p ${CONFIG_DIR}
    JWT_SECRET=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 64)
    CONFIG_FILE="${CONFIG_DIR}/pipelines.yaml"
    echo "Writing configuration to ${CONFIG_FILE}..."
    cat << EOF | sudo tee ${CONFIG_FILE}
server:
  host: 127.0.0.1
  port: 7000
  jwtSecret: ${JWT_SECRET}
  workers: 5
pipelines:
EOF
    sudo chown -R ${USER_NAME}:${USER_NAME} ${CONFIG_DIR}
fi

echo "Creating systemd service file..."
cat << EOF | sudo tee ${SYSTEMD_DIR}/${SERVICE_NAME}
[Unit]
Description=Micro CI Service
After=network.target

[Service]
User=${USER_NAME}
Group=${USER_NAME}
ExecStart=${INSTALL_DIR}/${PROGRAM_NAME}
WorkingDirectory=${CONFIG_DIR}
Restart=always
RestartSec=5
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload

if ! systemctl is-enabled --quiet ${SERVICE_NAME}; then
    echo "Enabling the service..."
    sudo systemctl enable ${SERVICE_NAME}
fi

echo "Starting the service..."
sudo systemctl restart ${SERVICE_NAME}

echo "${PROGRAM_NAME} installation is complete and the service is running."

