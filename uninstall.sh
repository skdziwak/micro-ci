#!/bin/bash

# Exit on any error
set -e

# Define the program name
PROGRAM_NAME="micro-ci"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="${PROGRAM_NAME}.service"
SYSTEMD_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/${PROGRAM_NAME}"
USER_NAME="micro-ci"

if [ "$1" == "--purge" ]; then
  echo "Purging configuration directory..."
  sudo rm -rf ${CONFIG_DIR}
fi

echo "Stopping the service if it's running..."
if systemctl is-active --quiet ${SERVICE_NAME}; then
  sudo systemctl stop ${SERVICE_NAME}
fi

echo "Disabling the service..."
if systemctl is-enabled --quiet ${SERVICE_NAME}; then
  sudo systemctl disable ${SERVICE_NAME}
fi

echo "Removing the service file..."
sudo rm -f ${SYSTEMD_DIR}/${SERVICE_NAME}

echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

echo "Removing the binary..."
sudo rm -f ${INSTALL_DIR}/${PROGRAM_NAME}

if id "$USER_NAME" &>/dev/null; then
    echo "Deleting the system user..."
    sudo userdel ${USER_NAME}
fi

echo "${PROGRAM_NAME} has been successfully uninstalled."
