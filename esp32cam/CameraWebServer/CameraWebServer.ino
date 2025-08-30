#include "esp_camera.h"
#include <WiFi.h>
#include <HTTPClient.h>
#include <WebSocketsClient.h>

//dane wifi
const char* ssid = "Orange_Swiatlowod_3060";
const char* password = "4X4y2NqTCpkf9U9Cdn";

const char* serverIp = "192.168.1.33"; // IP serwera
//const char* endpoint = "/api/camera?id=drzwi";
uint16_t port = 80;
const char* endpoint = "/api/camera?id=brama";

WebSocketsClient webSocket;

#define CAMERA_MODEL_AI_THINKER

#if defined(CAMERA_MODEL_AI_THINKER)
#define LED_FLASH 4  
#define PWDN_GPIO_NUM     32
#define RESET_GPIO_NUM    -1
#define XCLK_GPIO_NUM      0
#define SIOD_GPIO_NUM     26
#define SIOC_GPIO_NUM     27
#define Y9_GPIO_NUM       35
#define Y8_GPIO_NUM       34
#define Y7_GPIO_NUM       39
#define Y6_GPIO_NUM       36
#define Y5_GPIO_NUM       21
#define Y4_GPIO_NUM       19
#define Y3_GPIO_NUM       18
#define Y2_GPIO_NUM        5
#define VSYNC_GPIO_NUM    25
#define HREF_GPIO_NUM     23
#define PCLK_GPIO_NUM     22
#else
#error "Wybierz poprawny model kamery!"
#endif

void webSocketEvent(WStype_t type, uint8_t * payload, size_t length) {
    switch(type) {
        case WStype_DISCONNECTED:
            Serial.println("[WSc] Disconnected!\n");
            break;
        case WStype_CONNECTED:
            Serial.println("[WSc] Connected\n");
            break;
        case WStype_PING: 
            Serial.println("[WSc] Received ping");
            break;
        case WStype_PONG:  
            Serial.println("[WSc] Received pong");
            break;
        case WStype_TEXT:
        case WStype_BIN:
        case WStype_ERROR:		
            Serial.println("Error\n");	
            break;
        case WStype_FRAGMENT_TEXT_START:
        case WStype_FRAGMENT_BIN_START:
        case WStype_FRAGMENT:
        case WStype_FRAGMENT_FIN:
            break;
    }
}



void setup() {
  Serial.begin(115200);
  Serial.setDebugOutput(false);
  Serial.println();

  camera_config_t config;
  config.ledc_channel = LEDC_CHANNEL_0;
  config.ledc_timer   = LEDC_TIMER_0;
  config.pin_d0       = Y2_GPIO_NUM;
  config.pin_d1       = Y3_GPIO_NUM;
  config.pin_d2       = Y4_GPIO_NUM;
  config.pin_d3       = Y5_GPIO_NUM;
  config.pin_d4       = Y6_GPIO_NUM;
  config.pin_d5       = Y7_GPIO_NUM;
  config.pin_d6       = Y8_GPIO_NUM;
  config.pin_d7       = Y9_GPIO_NUM;
  config.pin_xclk     = XCLK_GPIO_NUM;
  config.pin_pclk     = PCLK_GPIO_NUM;
  config.pin_vsync    = VSYNC_GPIO_NUM;
  config.pin_href     = HREF_GPIO_NUM;
  config.pin_sscb_sda = SIOD_GPIO_NUM;
  config.pin_sscb_scl = SIOC_GPIO_NUM;
  config.pin_pwdn     = PWDN_GPIO_NUM;
  config.pin_reset    = RESET_GPIO_NUM;
  config.xclk_freq_hz = 20000000;
  config.pixel_format = PIXFORMAT_JPEG;

  

  // rozdzielczość i jakość
  config.frame_size = FRAMESIZE_VGA; // FRAMESIZE_SVGA, UXGA, QVGA, ...
  config.jpeg_quality = 8;           // 0-63 (niższa = lepsza jakość)
  config.fb_count = 1;

  pinMode(LED_FLASH, OUTPUT);  


  // Inicjalizacja kamery
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Błąd inicjalizacji kamery: 0x%x", err);
    digitalWrite(LED_FLASH, HIGH);
    return;
  }

  // Łączenie z Wi-Fi

  WiFi.begin(ssid, password);
  Serial.print("Łączenie z WiFi");
  while (WiFi.status() != WL_CONNECTED) {
    Serial.print(".");
    delay(1000);
  }
  Serial.println();
  Serial.println("Połączono z WiFi!");
  Serial.print("Adres IP: ");
  Serial.println(WiFi.localIP());

  webSocket.begin(serverIp, port, endpoint);
  webSocket.onEvent(webSocketEvent);
  webSocket.setReconnectInterval(5000);
  webSocket.enableHeartbeat(10000, 3000, 2);
}

const unsigned long timerDelay = 500;
const unsigned long timerWifi = 1000;
const unsigned long maxWifiAttempts = 30;

unsigned long lastTime = 0;
unsigned long lastWiFiCheck = 0;
unsigned int noWifiCounter = 0;

void CheckWifiConnection(){
  if (millis() - lastWiFiCheck > timerWifi) {
    if (WiFi.status() != WL_CONNECTED) {
      Serial.println("WiFi disconnected! Trying to reconnect...");
      WiFi.begin(ssid, password);
      noWifiCounter++;

      if (noWifiCounter >= maxWifiAttempts) { // Po 30 sekundach restart urzadzenia
        Serial.println("WiFi reconnect failed after 30 tries. Restarting ESP...");
        noWifiCounter = 0;
        ESP.restart();
      }
    } else {
      noWifiCounter = 0;
    }
    lastWiFiCheck = millis();
  }
}
void SendImage(){
  // Wysyłanie zdjęcia co 'timerDelay' ms
  if (millis() - lastTime > timerDelay) {
    if (WiFi.status() != WL_CONNECTED){
      Serial.println("WiFi not connected");
      return;
    }
    if(!webSocket.isConnected()){
      Serial.println("Websocket not connected");
      return;
    }

    camera_fb_t *fb = esp_camera_fb_get();
    if (!fb) {
      Serial.println("Failed to capture image");
      return;
    }
    bool success = webSocket.sendBIN(fb->buf, fb->len);
    if (!success) {
      Serial.println("Failed to send image");
    }
    esp_camera_fb_return(fb);
    lastTime = millis();
  }
    
}

void loop() {
  webSocket.loop();  
  SendImage();
  CheckWifiConnection();
}