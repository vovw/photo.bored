# Photo.Bored Deployment Guide

## 🚀 Quick Start

Photo.Bored is optimized for blazing fast performance and super easy deployment. Choose your preferred method below.

## Deployment Options

### 1. Docker (Recommended) 🐳

**Easiest way to deploy anywhere!**

```bash
# Start with Docker Compose
./deploy.sh docker

# Or manually:
docker-compose up -d
```

Access at: http://localhost:8080

**Custom port:**
```bash
PORT=3000 docker-compose up -d
```

### 2. VPS/Server Deployment 🖥️

**Deploy as a system service:**

```bash
# Run as root
sudo ./deploy.sh system
```

This will:
- Build the application
- Install to `/opt/photo-bored`
- Create systemd service
- Start automatically on boot

**Manage the service:**
```bash
sudo systemctl status photo-bored
sudo systemctl restart photo-bored
sudo systemctl stop photo-bored
sudo journalctl -u photo-bored -f  # View logs
```

### 3. Local Development 💻

```bash
# Build
./deploy.sh local

# Run
./app

# Or with custom config:
PORT=3000 DEBUG=true ./app
```

## Environment Configuration

Create a `.env` file (see `.env.example`):

```bash
PORT=8080                    # Server port
DB_PATH=./data/photos.db     # SQLite database path
UPLOAD_DIR=./uploads         # Upload directory
MAX_UPLOAD_SIZE=52428800     # Max upload size (50MB)
IMAGE_WORKERS=4              # Concurrent image processors
DEBUG=false                  # Debug mode
```

## Performance Features ⚡

- **Automatic image optimization** - Creates WebP, JPEG variants
- **Responsive images** - Thumbnail, medium, large sizes
- **Lazy loading** - Images load on-demand
- **Gzip compression** - Reduced bandwidth
- **Aggressive caching** - 1-year cache for images
- **Worker pools** - Concurrent image processing
- **SQLite WAL mode** - Better concurrency
- **EXIF metadata** - Automatic extraction

## Reverse Proxy Setup (Nginx)

```nginx
server {
    listen 80;
    server_name photos.example.com;

    client_max_body_size 50M;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## SSL with Let's Encrypt

```bash
sudo certbot --nginx -d photos.example.com
```

## Backup

**Important directories to backup:**
- `./data/` - SQLite database
- `./uploads/` - Original images

```bash
# Backup script
tar -czf backup-$(date +%Y%m%d).tar.gz data/ uploads/
```

## Resource Requirements

**Minimum:**
- 512MB RAM
- 1 CPU core
- 10GB disk

**Recommended:**
- 2GB RAM
- 2 CPU cores
- 50GB+ disk

## Monitoring

Health check endpoint: `GET /health`

```bash
curl http://localhost:8080/health
# Returns: {"status":"ok"}
```

## Upgrading

### Docker:
```bash
docker-compose pull
docker-compose up -d
```

### System service:
```bash
cd /path/to/photo-bored
git pull
sudo ./deploy.sh system
```

## Troubleshooting

**Port already in use:**
```bash
# Change port
PORT=3000 docker-compose up -d
```

**Permission issues:**
```bash
sudo chown -R 1000:1000 data/ uploads/
```

**View logs:**
```bash
# Docker
docker-compose logs -f

# System service
sudo journalctl -u photo-bored -f
```

## Security Recommendations

1. Use HTTPS (Let's Encrypt)
2. Set up firewall (UFW)
3. Regular backups
4. Keep system updated
5. Use strong server passwords
6. Consider fail2ban for brute force protection

## Performance Tuning

**Increase image workers for faster processing:**
```bash
IMAGE_WORKERS=8 ./app
```

**Adjust max upload size:**
```bash
MAX_UPLOAD_SIZE=104857600  # 100MB
```

## Support

For issues, check the logs first:
```bash
# Docker
docker-compose logs -f photo-bored

# System service
sudo journalctl -u photo-bored -f
```
