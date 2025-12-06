# 🔒 Security Camera System

Real-time video surveillance system using ESP32-CAM and Go web server.

https://github.com/user-attachments/assets/37ed7cb8-430b-42d5-8277-bd959632e53d

![86aa94ae-e85a-48ce-b732-a09719383f28](https://github.com/user-attachments/assets/a687173f-ba33-4acb-b651-203c552cae4e)

## 🏗️ Architecture

```
ESP32-CAM → WebSocket → Go Server → WebSocket → Browser
                           ↓                      ↓      
                        Manager              Real-time Video
                           ↓                    
                        AI Analysis            
                           ↓
            Save image to disk if motion detected
```

## 🧰 Tech Stack
- Orange PI RV2 - mini-computer as server
- ESP32‑CAM (Arduino) – video capture device programming in C
- Go (1.21+) – net/http, goroutines, mutex
- Gorilla WebSocket – bidirectional real-time communication
- GoCV - Go wrapper for OpenCV
- SSD Mobilenet v1 COCO - pre-trained object detection model
- HTML5/CSS3, vanilla JavaScript – frontend without frameworks

## 🔎 How It Works – Step by Step

### 1) Server Startup
1. Application starts HTTP server
2. Configuration and logger are initialized
3. Service manager is created (WebSocket hub, storage, optional AI) in `internal/services`
4. Routes are registered in `internal/routes.SetupRoutes()` and `AuthMiddleware` is applied

### 2) User Interface Access
1. User opens `/`. `dynamicHTMLHandler` maps path to appropriate HTML file in `static/` (e.g., `/` → `static/index.html`)
2. Browser loads CSS/JS from `/static/*` and initializes WebSocket connection to `/api/view` (viewer channel)
3. UI shows placeholders and connection status (online/offline), as well as camera activity indicators

### 3) Camera Frame Delivery
1. Camera connects via WebSocket to `/api/camera` and transmits JPEG frames. Server forwards them to manager
2. Resolution and frame rate are configured on camera side
3. If camera doesn't respond to ping signal, connection is considered dead; maximum message size is also set

### 4) Broadcast to Viewers
1. WebSocket Hub (`internal/services/websocket`) broadcasts to all clients connected to `/api/view` JSON text messages:
        `{ "camera": "<name>", "image": "<base64 JPEG>" }`
2. Frontend sets image `img.src = "data:image/jpeg;base64,<...>"`, hides placeholder and marks camera as active
3. UI timer marks camera as offline if no frame arrives for specified time (e.g., 10s)

### 5) AI and Storage
1. Each frame is checked for motion detection
2. If motion detected, object recognition is attempted via one of the threads (object recognition is multi-threaded for performance)
3. When object is successfully recognized, a red rectangle is drawn around it and image is sent to buffer service
4. Frames wait in queue which cyclically saves queued frames to disk
5. Saved images gallery available at `/api/pictures`, view: `/api/pictures/view`, clear: `/api/pictures/clear`

### 6) Gallery
1. Photo data such as date, time, object and camera are contained in filename
2. This allows easy extraction of this data from JPG files, giving gallery advanced filter functionality to easily find selected photos

### 7) Logs and Administration
1. Operation logs: `/logs/info`, `/logs/warning`, `/logs/error`; clear: `/logs/*/clear`
2. Authorization: `/auth/login` (GET/POST) and `/auth/logout`. Most routes are protected by `AuthMiddleware`
3. Log files are easily accessible from web interface

## 📁 Structure 

```
esp32cam
├── CameraWebServer/
│   └── CameraWebServer.ino   # ESP32-cam camera program
WebServer/
├── cmd/
│   └── server/
│       └── main.go           # Server entry point
├── internal/
│   ├── app/                  # Application initialization
│   ├── config/               # Configuration
│   ├── handlers/             # HTTP/WS handlers (gallery, login, logs, websockets)
│   ├── logger/               # Logger for writing events to files
│   ├── middleware/           # Authentication middleware
│   ├── routes/               # Route registration
│   └── services/
│       ├── ai/               # Motion detection and object recognition service, AI models
│       ├── storage/          # Service for saving files to disk
│       └── websocket/        # Service for handling websockets with viewers
        └── manager.go        # Service management, handler-service communication
├── static/                   # Frontend files
│   ├── index.html, login.html, logs.html, pictures.html
│   ├── css/*.css
│   └── js/*.js
├── logs/                     # Log files (info/warning/error)
├── go.mod, go.sum
└── .vscode/
```


## 🚀 Installation and Setup

### Prerequisites
- Go 1.21+
- ESP32-CAM board
- Arduino IDE or PlatformIO
- Wi-Fi network
- (Optional) Docker and Docker Compose

### 1. ESP32-CAM Configuration

```cpp
// In CameraWebServer.ino change:
const char* ssid = "YourWiFi";
const char* password = "YourPassword";
const char* serverURL = "http://SERVER_IP:PORT/upload";
// Additionally, for different cameras add different query as 'camera' id e.g. /api/camera?id=door
```

### 2. Password Setup
1. Create `.env` file in WebServer folder
2. Add password and port as follows:
```env
PASSWORD=example_password
PORT=80
```

### 3. Running Go Server (Native)

```bash
cd WebServer
go mod tidy
go run cmd/server/main.go
```

Server will start on `http://localhost:80` (or port specified in `.env`)

### 4. Upload Code to ESP32-CAM

1. Open `esp32cam/CameraWebServer.ino` in Arduino IDE
2. Select board "AI Thinker ESP32-CAM" 
3. Upload code to ESP32-CAM

## 🐳 Docker Setup

The project includes pre-configured Docker files in `WebServer/` directory:
- `docker-compose.yml` - Docker Compose configuration
- `Dockerfile` - Multi-stage build with OpenCV 4.10.0 support

### Quick Start with Docker Compose

1. **Configure environment variables:**

Create `.env` file in `WebServer/` directory:
```env
PASSWORD=your_secure_password
PORT=8080
HOST_PORT=8080
PROCESSING_WORKERS=4
```

2. **Start the application:**

```bash
cd WebServer
docker-compose up -d
```

3. **Access the application:**
- Web interface: `http://localhost:8080`
- WebSocket (cameras): `ws://localhost:8080/api/camera?id=camera_name`
- WebSocket (viewers): `ws://localhost:8080/api/view`

4. **Manage containers:**

```bash
# View logs
docker-compose logs -f webserver

# Stop services
docker-compose down

# Rebuild after code changes
docker-compose up -d --build
```

### Docker Configuration Details

**`docker-compose.yml` features:**
- Port mapping: `${HOST_PORT:-8080}:${PORT:-8080}`
- Environment variables with defaults
- Persistent volumes for `/static` and `/logs`
- Auto-restart policy (`unless-stopped`)
- Isolated bridge network

**`Dockerfile` features:**
- Multi-stage build (golang:1.21-bookworm → debian:bookworm-slim)
- OpenCV 4.10.0 compiled from source with GoCV support
- All AI models and static files included
- CGO enabled for OpenCV bindings
- Optimized runtime image (~2GB)

