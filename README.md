# 🔒 System Kamer Bezpieczeństwa

System monitoringu wideo w czasie rzeczywistym wykorzystujący ESP32-CAM i serwer webowy napisany w Go.


## 🏗️ Architektura

```
ESP32-CAM → WebSocket → Go Server → WebSocket → Browser
                           ↓                      ↓      
                        Manager              Real-time Video
                           ↓                    
                       Analiza AI            
                           ↓
            Zapisanie obrazu na dysku jeśli wykryto ruch
```

## 🧰 Tech stack
- Orange PI RV2 - minikomputer jako serwer
- ESP32‑CAM (Arduino) – programowanie urządzenia rejestrującego obraz w C
- Go (1.21+) – net/http, goroutines, mutex
- Gorilla WebSocket – dwukierunkowa komunikacja w czasie rzeczywistym
- GoCV - nakładka do GO która umożliwia korzystanie z OpenCV
- SSD Mobilenet v1 coco - gotowy model do rozpoznawania obiektów
- HTML5/CSS3, czysty JavaScript – frontend bez frameworków

## 🔎 Jak to działa – krok po kroku

### 1) Start serwera
1. Aplikacja uruchamia serwer HTTP
2. Inicjalizowana jest konfiguracja oraz logger.
3. Tworzony jest menedżer usług (hub WebSocket, storage, opcjonalnie AI) w `internal/services`.
4. Rejestrowane są trasy w `internal/routes.SetupRoutes()` i nakładany `AuthMiddleware`.

### 2) Wejście użytkownika do UI
1. Użytkownik otwiera `/`. `dynamicHTMLHandler` mapuje ścieżkę na odpowiedni plik HTML w `static/` (np. `/` → `static/index.html`).
2. Przeglądarka ładuje CSS/JS z `/static/*` i inicjuje połączenie WebSocket do `/api/view` (kanał dla oglądających).
3. UI pokazuje placeholdery i status połączenia (online/offline), a także wskaźniki aktywności kamer.

### 3) Dostarczanie ramek z kamer
1. Kamera łączy się WebSocketem na `/api/camera` i przesyła ramki JPEG. Serwer przekazuje je do managera.
2. Rozdzielczość i ilość klatek na sekunde ustawiania jest po stronie kamery.
3. Jeśli kamera nie odpowie na sygnał ping to połączenie uznawane jest za martwe, ustawiona jest również maksymalna wielkość wiadomości

### 4) Broadcast do oglądających
1. Hub WebSocket (`internal/services/websocket`) rozsyła do wszystkich klientów podłączonych do `/api/view` wiadomości tekstowe JSON:
        `{ "camera": "<nazwa>", "image": "<base64 JPEG>" }`.
2. Frontend ustawia obraz `img.src = "data:image/jpeg;base64,<...>"`, ukrywa placeholder i oznacza kamerę jako aktywną.
3. Timer w UI oznacza kamerę jako offline, jeśli przez określony czas (np. 10 s) nie dotarła żadna ramka.

### 5) AI i zapisy 
1. Każda klatka jest sprawdzana pod kątem wykrycia na niej ruchu
2. Jeśli wykryto ruch to następuje próba rozpoznania obiektu na obrazie poprzez jeden z wątków (rozpoznawanie obiektu jest podzielone na wątki, aby przyspieszyć działanie)
3. Gdy poprawnie rozpoznano obiekt, wokół niego rysowany jest czerwony prostokąt, a obraz wysyłany jest do buffer serwisu
4. Tam ramki czekają w kolejce, która cyklicznie znajdujące się w niej klatki, zapisuje na dysku
5. Przegląd zapisów dostępny jest pod `/api/pictures`, podgląd: `/api/pictures/view`, czyszczenie: `/api/pictures/clear`.

### 6) Galeria
1. Dane zdjęcia takie jak data, godzina, obiekt i kamera są zawarte w nazwie pliku
2. Umożliwia to łatwe wyłuskanie tych danych z plików jpg, dzięki czemu galeria ma rozbudowanę funkcję filtrów, które ułatwiają znalezienie wybranych fotografii

### 7) Logi i administracja
1. Logi działania: `/logs/info`, `/logs/warning`, `/logs/error`; czyszczenie: `/logs/*/clear`.
2. Autoryzacja: `/auth/login` (GET/POST) i `/auth/logout`. Większość tras jest chroniona przez `AuthMiddleware`.
3. Pliki z logami są łatwo dostepne ze strony internetowej


## 📁 Struktura 

```
esp32cam
├── CameraWebServer/
│   └── CameraWebServer.ino   # Program kamery esp32-cam
WebServer/
├── cmd/
│   └── server/
│       └── main.go           # Punkt wejścia serwera
├── internal/
│   ├── app/                  # Inicjalizacja aplikacji
│   ├── config/               # Konfiguracja
│   ├── handlers/             # Handlery HTTP/WS (gallery, login, logs, websockets)
│   ├── logger/               # Logger do zapisywania wydarzeń do plików
│   ├── middleware/           # Middleware do autoryzacji
│   ├── routes/               # Rejestracja tras
│   └── services/
│       ├── ai/               # Serwis do wykrywanie ruchu i rozpoznawanie obiektów, modele AI
│       ├── storage/          # Serwis do zapisywania plików na dysk
│       └── websocket/        # Serwis do obsługi websocketów z oglądającymi
        └── manager.go        # Zarządzanie serwisami, komunikacja handlerow z serwisami
├── static/                   # Pliki frontend
│   ├── index.html, login.html, logs.html, pictures.html
│   ├── css/*.css
│   └── js/*.js
├── logs/                     # Pliki z logami (info/warning/error)
├── go.mod, go.sum
└── .vscode/
```


## 🚀 Instalacja i Uruchomienie

### 1. Konfiguracja ESP32-CAM

```cpp
// W CameraWebServer.ino zmień:
const char* ssid = "TwojaWiFi";
const char* password = "TwojeHaslo";
const char* serverURL = "http://IP_SERWERA:PORT/upload";
// Ponadto dla różnych kamer dodaj różne query jako w 'camera' jako id np /api/camera?id=drzwi

```
### 2. Ustawienie hasła
1. Utwórz plik .env w folderze Webserver
2. Umieść tam hasło w następujący sposób
```
PASSWORD=example_password
```


### 3. Uruchomienie serwera Go

```bash
cd WebServer
go mod tidy
go run cmd/server/main.go
```

### 3. Wgranie kodu na ESP32-CAM

1. Otwórz `esp32cam/CameraWebServer.ino` w Arduino IDE
2. Wybierz board "AI Thinker ESP32-CAM" 
3. Wgraj kod na ESP32-CAM


