#include "esp_camera.h"
#include <WiFi.h>
#include <HTTPClient.h>
#include <WebSocketsClient.h>

//dane wifi
const char* ssid = "Orange_Swiatlowod_3060";
const char* password = "4X4y2NqTCpkf9U9Cdn";

const char* serverIp = "192.168.1.33"; // IP serwera
const char* endpoint = "/api/camera?id=drzwi";
uint16_t port = 8080;
//const char* endpoint = "/api/camera?id=balkon";

WebSocketsClient webSocket;

#define CAMERA_MODEL_AI_THINKER

#if defined(CAMERA_MODEL_AI_THINKER)
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
		case WStype_TEXT:
		case WStype_BIN:
		case WStype_ERROR:			
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
  config.jpeg_quality = 6;           // 0-63 (niższa = lepsza jakość)
  config.fb_count = 1;

  // Inicjalizacja kamery
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Błąd inicjalizacji kamery: 0x%x", err);
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
}

unsigned long lastTime = 0;
unsigned long timerDelay = 250;

void loop() {
  webSocket.loop();
  if(millis() - lastTime > timerDelay){
    camera_fb_t *fb = esp_camera_fb_get();
    if (!fb) return;
    
    if (WiFi.status() == WL_CONNECTED) {
      // String url = String(serverAdress) + String(endpoint);
      // http.begin(url);
      // http.addHeader("Content-Type", "image/jpeg");
      // int httpResponseCode = http.POST(fb->buf, fb->len);
      // Serial.println(httpResponseCode);
      // http.end();
      // ^--przesylanie po http 
      if(webSocket.isConnected()){
        webSocket.sendBIN(fb->buf, fb->len);
      }
      else{
        Serial.println("Problem with connecting websocket");
      }
      
    }
    esp_camera_fb_return(fb);
    lastTime = millis();
  }
 
}