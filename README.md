# Micro-CI

Micro-CI is a compact, lightweight Continuous Integration (CI) server aimed at developers who wish to self-manage their servers and effortlessly host their personal projects. It is designed to run efficiently on limited memory resources, making it especially suitable for lower-end virtual private servers. While Micro-CI fulfills basic CI duties with ease, it is intended as a straightforward tool rather than a full-fledged production system—a role better served by more comprehensive systems like Jenkins.

## Features

- **Lightweight and Efficient**: Optimally designed for small-scale environments with limited resources.
- **Simple Configuration**: Set up pipelines with an easy-to-understand YAML file.
- **Secure Webhook Execution with JWT**: Pipelines are triggered via secure POST requests using JWT tokens for authenticity verification.
- **Concurrent Execution**: Manage task execution in parallel with configurable workers.
- **Event-driven Notifications**: Integrate with messaging systems to notify stakeholders of task outcomes.
- **SSH Key Management**: Automatically generate and use SSH keys for private repository cloning.
- **Temp Directory Cleanup**: Manage the creation and deletion of temp directories for each job.

## Installation

Micro-CI can be installed using the provided `install.sh` script, which will compile the binary, establish the necessary system user, and set up Micro-CI as a systemd service. To install or rebuild Micro-CI, run the following command:

```bash
./install.sh --rebuild
```

## Configuration

A `pipelines.yaml` file is used to setup Micro-CI. Here is an example of a server configuration with one pipeline:

```yaml
server:
  host: 127.0.0.1
  port: 7000
  jwtSecret: secret
  notificationCommand: |
    curl -X POST https://api.telegram.org/bot<telegram_token>/sendMessage -d chat_id=<chat_id> -d text="$MESSAGE"
  workers: 1
pipelines:
  - name: bridge
    repository: git@github.com:skdziwak/ts3-telegram-bridge-v2.git
    script: |
      set -e
      cp /etc/micro-ci/configs/bridge/config.yaml .
      docker build --tag ts3-telegram-bridge .
      docker rm -f ts3-telegram-bridge || true
      docker run --name ts3-telegram-bridge --network mongonet --restart always -d ts3-telegram-bridge
```

Make sure to replace `<telegram_token>` and `<chat_id>` with actual values. The `target` states which branch or commit the pipeline should checkout.

## Running the Service

Upon installation, Micro-CI operates as a system service. Control it using standard systemd commands:

```bash
# Start the service:
sudo systemctl start micro-ci.service

# Stop the service:
sudo systemctl stop micro-ci

# Restart the service:
sudo systemctl restart micro-ci

# Service status:
sudo systemctl status micro-ci
```

## Usage

Triggers for pipelines are set by JWT-secured POST requests:

```bash
curl -X POST http://127.0.0.1:7000/bridge?token=<your_jwt_token>
```

You can generate a JWT with:

```bash
micro-ci jwt
```

This produces a JWT based on the server-defined secret, which should be used as the path parameter in webhook triggers.

## Considerations

Micro-CI efficiently handles CI tasks for personal or small-scale projects but does not offer the extensive features of more complex systems. It's excellent for simple use cases but not designed for production-level demands.
