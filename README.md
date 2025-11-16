# photo.bored ⚡

A blazingly fast, self-hosted image blogging platform. Simple, performant, and easy to deploy.

## ✨ Features

- 🚀 **Blazingly fast** - Optimized image processing with WebP support
- 📱 **Responsive images** - Automatic thumbnails and multiple sizes
- 💾 **Smart caching** - Aggressive caching for optimal performance
- 🔄 **Lazy loading** - Images load on-demand
- 🎨 **Clean UI** - Non-cluttered, Instagram-inspired design
- 📊 **EXIF metadata** - Automatic date/location extraction
- 🐳 **Easy deployment** - Docker, VPS, or local
- 💯 **Free forever** - Self-host on your infrastructure

## 🚀 Quick Start

### Docker (Recommended)
```bash
docker-compose up -d
```

### Local Development
```bash
go build -o app
./app
```

### VPS Deployment
```bash
sudo ./deploy.sh system
```

## 📖 Documentation

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## 🎯 Goals

- ✅ Organize photos by albums (lists)
- ✅ Clean, minimal UI
- ✅ Easy to self-host
- ✅ Fast performance
- ✅ Automatic image optimization

## 🚫 Anti-Goals

- ❌ Ads
- ❌ Tracking
- ❌ Cluttered interface
- ❌ Complex setup

## 🛠️ Tech Stack

- **Backend:** Go (Golang)
- **Database:** SQLite with WAL mode
- **Image Processing:** Concurrent workers with WebP support
- **Frontend:** Vanilla JS (no frameworks)
- **Deployment:** Docker, systemd

## 📁 Project Structure

```
├── main.go        - Server setup & routing
├── site.go        - HTTP handlers
├── sqlite.go      - Database operations
├── imageproc.go   - Image processing & optimization
├── middleware.go  - Gzip, caching, security
├── config.go      - Environment configuration
├── index.html     - Frontend UI
└── deploy.sh      - Deployment script
```

## 🔧 Configuration

Environment variables:
- `PORT` - Server port (default: 8080)
- `DB_PATH` - Database path (default: ./data/photos.db)
- `UPLOAD_DIR` - Upload directory (default: ./uploads)
- `MAX_UPLOAD_SIZE` - Max upload size in bytes (default: 50MB)
- `IMAGE_WORKERS` - Concurrent processors (default: 4)
- `DEBUG` - Debug mode (default: false)

## 📊 Performance

- Automatic WebP conversion (30-50% smaller)
- Multiple image sizes (thumb, medium, large)
- 1-year browser cache for images
- Gzip compression for responses
- Concurrent image processing
- SQLite WAL mode for better concurrency

## 🙏 License

Free and open source. Do whatever you want with it!
