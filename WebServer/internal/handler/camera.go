package handler

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"webserver/internal/config"
	"webserver/internal/logger"
	"webserver/internal/service"
)

var (
	jpegHeader = []byte{0xFF, 0xD8}
	jpegFooter = []byte{0xFF, 0xD9}
)

// UDPCameraHandler listens for UDP packets from cameras, reconstructs JPEG frames,
// and forwards complete frames to the Manager for processing.
func UDPCameraHandler(manager *service.Manager, logger *logger.Logger, config *config.Config) {
	port := strconv.Itoa(config.CamerasPort)

	addr, err := net.ResolveUDPAddr("udp", ":"+port)

	if err != nil {
		logger.Error("Failed to resolve UDP address: %v", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("Failed to listen on UDP port %s: %v", port, err)
		return
	}
	defer conn.Close()

	logger.Info("UDP Camera handler started on port %s", port)
	buffer := make([]byte, 2048)

	cameraBuffers := make(map[string]*bytes.Buffer)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("Error reading UDP packet: %v", err)
			continue
		}

		ip := strings.Split(remoteAddr.String(), ":")[0]
		cameraName, exists := config.CameraNames[ip]
		if !exists {
			cameraName = "unknown_" + ip
		}

		data := buffer[:n]
		if _, ok := cameraBuffers[cameraName]; !ok {
			cameraBuffers[cameraName] = new(bytes.Buffer)
		}
		imgBuffer := cameraBuffers[cameraName]

		if bytes.HasPrefix(data, jpegHeader) {
			imgBuffer.Reset()
		}
		imgBuffer.Write(data)

		if bytes.HasSuffix(data, jpegFooter) {
			fullFrame := make([]byte, imgBuffer.Len())
			copy(fullFrame, imgBuffer.Bytes())
			manager.HandleCameraImage(fullFrame, cameraName)
			imgBuffer.Reset()
		}
	}
}
