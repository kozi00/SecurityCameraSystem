# Build stage
FROM golang:1.21-bookworm AS builder

# Install OpenCV dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config \
    wget \
    unzip \
    libgtk-3-dev \
    libavcodec-dev \
    libavformat-dev \
    libswscale-dev \
    libv4l-dev \
    libxvidcore-dev \
    libx264-dev \
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    gfortran \
    openexr \
    libatlas-base-dev \
    libtbb12 \
    libtbb-dev \
    libdc1394-dev \
    libopenexr-dev \
    libgstreamer-plugins-base1.0-dev \
    libgstreamer1.0-dev \
    && rm -rf /var/lib/apt/lists/*

# Install OpenCV
ENV OPENCV_VERSION=4.10.0
RUN cd /tmp && \
    wget -O opencv.zip https://github.com/opencv/opencv/archive/${OPENCV_VERSION}.zip && \
    wget -O opencv_contrib.zip https://github.com/opencv/opencv_contrib/archive/${OPENCV_VERSION}.zip && \
    unzip opencv.zip && \
    unzip opencv_contrib.zip && \
    mkdir -p opencv-${OPENCV_VERSION}/build && \
    cd opencv-${OPENCV_VERSION}/build && \
    cmake -D CMAKE_BUILD_TYPE=RELEASE \
          -D CMAKE_INSTALL_PREFIX=/usr/local \
          -D OPENCV_EXTRA_MODULES_PATH=/tmp/opencv_contrib-${OPENCV_VERSION}/modules \
          -D WITH_GSTREAMER=ON \
          -D WITH_FFMPEG=ON \
          -D BUILD_DOCS=OFF \
          -D BUILD_EXAMPLES=OFF \
          -D BUILD_TESTS=OFF \
          -D BUILD_PERF_TESTS=OFF \
          -D BUILD_opencv_java=OFF \
          -D BUILD_opencv_python=OFF \
          -D BUILD_opencv_python2=OFF \
          -D BUILD_opencv_python3=OFF \
          .. && \
    make -j$(nproc) && \
    make install && \
    ldconfig && \
    cd / && rm -rf /tmp/opencv*

# Create pkg-config file for OpenCV
RUN mkdir -p /usr/local/lib/pkgconfig && \
    echo "prefix=/usr/local" > /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "exec_prefix=\${prefix}" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "libdir=\${exec_prefix}/lib" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "includedir=\${prefix}/include/opencv4" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "Name: OpenCV" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "Description: Open Source Computer Vision Library" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "Version: ${OPENCV_VERSION}" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "Libs: -L\${libdir} -lopencv_core -lopencv_imgproc -lopencv_imgcodecs -lopencv_videoio -lopencv_highgui -lopencv_video -lopencv_objdetect -lopencv_dnn -lopencv_features2d -lopencv_flann -lopencv_calib3d -lopencv_photo" >> /usr/local/lib/pkgconfig/opencv4.pc && \
    echo "Cflags: -I\${includedir}" >> /usr/local/lib/pkgconfig/opencv4.pc

ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/server/main.go

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies for OpenCV
RUN apt-get update && apt-get install -y \
    libgtk-3-0 \
    libavcodec59 \
    libavformat59 \
    libswscale6 \
    libv4l-0 \
    libxvidcore4 \
    libx264-164 \
    libjpeg62-turbo \
    libpng16-16 \
    libtiff6 \
    libopenexr-3-1-30 \
    libatlas-base-dev \
    libtbb12 \
    libdc1394-25 \
    libgstreamer-plugins-base1.0-0 \
    libgstreamer1.0-0 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy OpenCV from builder
COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /usr/local/include /usr/local/include
COPY --from=builder /usr/local/share/opencv4 /usr/local/share/opencv4
COPY --from=builder /usr/local/lib/pkgconfig /usr/local/lib/pkgconfig

# Update library cache
RUN ldconfig

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server /app/server

# Copy necessary files
COPY --from=builder /app/internal/services/ai/*.pb /app/internal/services/ai/
COPY --from=builder /app/internal/services/ai/*.pbtxt /app/internal/services/ai/
COPY --from=builder /app/static /app/static

# Create directories for logs and images
RUN mkdir -p /app/logs /app/static/images

# Expose port (default 8080)
EXPOSE 8080

# Set environment variables with defaults
ENV PORT=8080
ENV PASSWORD=""
ENV STATIC_DIR="/app/static"
ENV MODEL_PATH="/app/internal/services/ai/frozen_inference_graph.pb"
ENV CONFIG_PATH="/app/internal/services/ai/ssd_mobilenet_v1_coco_2017_11_17.pbtxt"
ENV IMAGE_DIR="/app/static/images"
ENV LOG_DIR="/app/logs"
ENV PROCESSING_WORKERS=4

# Run the application
CMD ["/app/server"]
