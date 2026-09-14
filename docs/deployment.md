# Production Deployment Guide

## Standalone Process Deployment

1. **Deploy Binary & Config:**
   Copy `ekarouter.exe` (or `ekarouter` on Linux) and `.env` to the target directory:
   ```text
   /opt/ekarouter/
     ├── ekarouter
     ├── .env
     ├── migrations/
     └── data/
   ```

2. **Systemd Service (Linux):**
   `/etc/systemd/system/ekarouter.service`:
   ```ini
   [Unit]
   Description=EkaRouter Universal Developer Platform
   After=network.target

   [Service]
   Type=simple
   User=ekarouter
   WorkingDirectory=/opt/ekarouter
   ExecStart=/opt/ekarouter/ekarouter serve
   Restart=always
   RestartSec=5
   EnvironmentFile=/opt/ekarouter/.env

   [Install]
   WantedBy=multi-user.target
   ```

3. **Windows Service:**
   Use NSSM (Non-Sucking Service Manager) or Windows Task Scheduler to run `ekarouter.exe serve` on system startup.
