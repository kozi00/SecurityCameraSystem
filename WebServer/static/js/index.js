class CameraMonitor {
    constructor() {
        this.socket = null;
        this.connectionBadge = document.getElementById('connection-badge');
        this.cameraStatus = {
            balkon: document.getElementById('status_balkon'),
            drzwi: document.getElementById('status_drzwi')
        };
        this.cameraActivity = {
            balkon: Date.now(),
            drzwi: Date.now()
        };
        
        this.init();
    }

    init() {
        this.connectWebSocket();
        this.startActivityChecker();
    }

    connectWebSocket() {
        this.socket = new WebSocket("ws://" + location.host + "/api/view");

        this.socket.onmessage = (event) => this.handleMessage(event);
        this.socket.onopen = (event) => this.handleOpen(event);
        this.socket.onclose = (event) => this.handleClose(event);
        this.socket.onerror = (error) => this.handleError(error);
        
    }

    handleMessage(event) {
        try {
            const data = JSON.parse(event.data);
            const base64Image = data.image;
            const camera = data.camera;

            
            if (!base64Image || !camera) {
                console.error("Brak wymaganych danych w wiadomości");
                return;
            }
            
            this.showImage(camera, "data:image/jpeg;base64," + base64Image);
            
        } catch (error) {
            console.error("Błąd podczas parsowania wiadomości WebSocket:", error);
        }
    }

    handleOpen(event) {
        console.log("Połączenie WebSocket nawiązane");
        this.updateConnectionStatus(true);
    }

    handleClose(event) {
        console.log("Połączenie WebSocket zamknięte");
        this.updateConnectionStatus(false);
        
        Object.keys(this.cameraStatus).forEach(camera => {
            this.updateCameraStatus(camera, false);
        });
    }

    handleError(error) {
        console.error("Błąd WebSocket:", error);
        this.updateConnectionStatus(false);
    }

    updateConnectionStatus(isConnected) {
        if (isConnected) {
            this.connectionBadge.className = 'status-badge connected';
            this.connectionBadge.innerHTML = '<div class="status-dot"></div><span>Połączono z serwerem</span>';
        } else {
            this.connectionBadge.className = 'status-badge disconnected';
            this.connectionBadge.innerHTML = '<div class="status-dot"></div><span>Rozłączono z serwerem</span>';
        }
    }

    updateCameraStatus(camera, isActive) {
        if (this.cameraStatus[camera]) {
            this.cameraStatus[camera].className = isActive ? 'status-indicator online' : 'status-indicator';
        }
    }

    showImage(camera, src) {
        const img = document.getElementById("camera_" + camera);
        const placeholder = img.parentNode.querySelector('.camera-placeholder');
        
        if (img && placeholder) {
            img.onload = () => {
                placeholder.style.display = 'none';
                img.style.display = 'block';
                img.style.opacity = '1';
            };
            img.src = src;
            
            this.cameraActivity[camera] = Date.now();
            this.updateCameraStatus(camera, true);
        }
    }
    startActivityChecker() {
        setInterval(() => {
            const now = Date.now();
            Object.keys(this.cameraActivity).forEach(camera => {
                const isActive = (now - this.cameraActivity[camera]) < 10000; 
                this.updateCameraStatus(camera, isActive);
            });
        }, 5000);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new CameraMonitor();
});
