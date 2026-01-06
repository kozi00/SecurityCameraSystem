#  Security Camera System

Real-time video surveillance system using ESP32-CAM and Go web server.

https://github.com/user-attachments/assets/af57837d-b1cb-4756-a829-995996c15331

![86aa94ae-e85a-48ce-b732-a09719383f28](https://github.com/user-attachments/assets/a687173f-ba33-4acb-b651-203c552cae4e)

##  Architecture

```
ESP32-CAM → UDP → Go Server → WebSocket → Browser
                           ↓                      ↓      
                        Manager              Real-time Video
                           ↓                    
                        AI Analysis            
                           ↓
            Save image to disk if motion detected
```

## Tech Stack
- **Hardware:** Orange PI RV2 (Server), ESP32‑CAM (Camera)
- **Camera Firmware:** C++ (Arduino IDE), UDP protocol for video streaming
- **Backend:** Go (1.21+) – `net` (UDP), `net/http`, goroutines, mutex
- **Communication:** - **Ingress:** UDP (Low latency video streaming from cameras)
  - **Egress:** Gorilla WebSocket (Real-time broadcasting to clients)
- **Computer Vision:** GoCV (OpenCV wrapper), SSD Mobilenet v1 COCO
- **Database**: SQLite, light and fast database engine
- **Frontend:** HTML5/CSS3, vanilla JavaScript
- **Infrastructure:** Docker, Docker Compose

## How It Works – Step by Step

### 1) Server Startup
1. Application starts HTTP server for viewers and UDP listener for cameras.
2. Configuration (`.env`) is loaded, mapping Camera IPs to friendly names (e.g., "Door", "Gate").
3. Service manager is created (WebSocket hub, storage, AI, UDP handler) in `internal/services`.
4. Routes are registered in `internal/routes.SetupRoutes()`.

### 2) User Interface Access
1. User opens `/`. `dynamicHTMLHandler` maps path to appropriate HTML file.
2. Browser loads CSS/JS and initializes WebSocket connection to `/api/view`.
3. UI shows placeholders and connection status.

### 3) Camera Frame Delivery (UDP)
1. **Fragmentation:** ESP32-CAM captures a JPEG frame and splits it into small chunks (max ~1436 bytes) to fit within network MTU.
2. **Streaming:** Chunks are sent via UDP to the server's specific port (e.g., 81).
3. **Reassembly:** Go Server listens on the UDP port. It identifies the camera by source IP address.
4. **Buffering:** Server reassembles chunks into a full JPEG frame. Once a valid start (0xFF, 0xD8) and end (0xFF, 0xD9) markers are found, the frame is passed to the Manager.

### 4) Broadcast to Viewers
1. WebSocket Hub broadcasts the complete JPEG frame to all connected clients as a JSON message:
        `{ "camera": "<name>", "image": "<base64 JPEG>" }`
2. Frontend updates the `img.src` attribute with the Base64 data.
3. UI timer marks camera as offline if no frame arrives for a specified time.

### 5) AI and Storage
1. Each assembled frame is passed to the motion detection service.
2. If motion is detected, the frame is queued for AI Object Recognition (multi-threaded).
3. If an object is recognized (e.g., Person, Car), a bounding box is drawn, and the image is saved to disk via `bufferService`.
4. Saved images are available in the gallery (`/api/pictures`).

### 6) Gallery & Logs
1. Filenames contain metadata (date, time, object, camera).
2. Advanced filtering available in the UI.
3. Logs accessible via `/logs/*` endpoints.

### 7) Database
1. Every image is also represented in SQLite relational database.
2. When saving image, it's data is also stored in Image and Detection tables.
3. It allows for better data analysis and management.

##  Structure 

```
esp32cam
├── CameraWebServer/
│   └── CameraWebServer.ino   # ESP32-cam camera program
WebServer/
├── cmd/
│   └── server/
│       └── main.go           # Server entry point
├── data/                     # SQLite files 
├── tests/                    # Integration tests
├── internal/
│   ├── app/                  # Application initialization
│   ├── config/               # Configuration
│   ├── handler/             # HTTP/WS handlers (gallery, login, logs, websockets)
│   ├── logger/               # Logger for writing events to files
│   ├── middleware/           # Authentication middleware
│   ├── route/               # Route registration
│   ├── model/               # Entities represented in database
│   ├── dto/                 # Data objects used for transfering information
│   ├── repository/          # Connecting to database and executing queries
│   └── service/
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


##  Installation and Setup

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
// Server Configuration (UDP Target)
// Note: Use commas, not dots for IPAddress
IPAddress serverIp(192, 168, 1, 10); 
const uint16_t udpPort = 81;          // Must match CAMERAS_PORT in .env
```

### 2. Password Setup
1. Create `.env` file in WebServer folder
2. Add variables as follows:
```env
# Server Auth and Web Port
PASSWORD=password123
PORT=8080

# UDP Configuration
CAMERAS_PORT=81
# Format: IP:Name,IP:Name
CAMERAS="192.168.1.32:Gate,192.168.1.33:Door"

# Performance
PROCESSING_WORKERS=4
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

##  Docker Setup

The project includes pre-configured Docker files in `WebServer/` directory:
- `docker-compose.yml` - Docker Compose configuration
- `Dockerfile` - Multi-stage build with OpenCV 4.10.0 support

### Quick Start with Docker Compose

1. **Configure environment variables:**

Create `.env` file in `WebServer/` directory:
```env
PASSWORD=password123
PORT=8080
CAMERAS_PORT=81
CAMERAS="192.168.1.32:Gate,192.168.1.33:Door"
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

