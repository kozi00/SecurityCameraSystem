# Download SSD MobileNet v2 Lite - lightweight model for object detection
# Perfect for security camera systems

Write-Host "🚀 Downloading SSD MobileNet v2 Lite..." -ForegroundColor Cyan

# Create models directory
$modelsDir = "internal\services\AI\models"
if (!(Test-Path $modelsDir)) {
    New-Item -ItemType Directory -Force -Path $modelsDir
    Write-Host "📁 Created models directory: $modelsDir" -ForegroundColor Green
}

# Download SSD MobileNet v2 Lite ONNX (best for GoCV)
Write-Host "📦 Downloading SSD MobileNet v2 Lite ONNX..." -ForegroundColor Yellow

$modelUrl = "https://github.com/onnx/models/raw/main/vision/object_detection_segmentation/ssd-mobilenetv1/model.onnx"
$modelFile = "$modelsDir\ssd_mobilenet_v2_lite.onnx"

try {
    Invoke-WebRequest -Uri $modelUrl -OutFile $modelFile -UseBasicParsing
    Write-Host "✅ SSD MobileNet v2 Lite downloaded!" -ForegroundColor Green
    
    $sizeMB = [math]::Round((Get-Item $modelFile).Length / 1MB, 2)
    Write-Host "📊 Model size: $sizeMB MB" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed to download: $_" -ForegroundColor Red
    Write-Host "🔄 Trying alternative source..." -ForegroundColor Yellow
    
    # Alternative download
    $altUrl = "https://storage.googleapis.com/tfhub-modules/tensorflow/ssd_mobilenet_v2/2.tar.gz"
    $altFile = "$modelsDir\ssd_mobilenet_v2_alternative.tar.gz"
    
    try {
        Invoke-WebRequest -Uri $altUrl -OutFile $altFile -UseBasicParsing
        Write-Host "✅ Alternative model downloaded!" -ForegroundColor Green
    } catch {
        Write-Host "❌ All downloads failed" -ForegroundColor Red
    }
}

# Download YOLOv5 nano as ultra-light alternative
Write-Host ""
Write-Host "🎯 Downloading YOLOv5 nano (ultra-lightweight ~1.9MB)..." -ForegroundColor Cyan

$yoloUrl = "https://github.com/ultralytics/yolov5/releases/download/v7.0/yolov5n.onnx"
$yoloFile = "$modelsDir\yolov5n_lite.onnx"

try {
    Invoke-WebRequest -Uri $yoloUrl -OutFile $yoloFile -UseBasicParsing
    Write-Host "✅ YOLOv5 nano downloaded!" -ForegroundColor Green
    
    $sizeMB = [math]::Round((Get-Item $yoloFile).Length / 1MB, 2)
    Write-Host "📊 Model size: $sizeMB MB" -ForegroundColor Gray
} catch {
    Write-Host "❌ Failed to download YOLOv5: $_" -ForegroundColor Red
}

# Manual download instructions
Write-Host ""
Write-Host "🔗 Manual download options (if above failed):" -ForegroundColor Yellow
Write-Host "1. SSD MobileNet v2:" -ForegroundColor Gray
Write-Host "   https://tfhub.dev/tensorflow/ssd_mobilenet_v2/2" -ForegroundColor White
Write-Host "2. TensorFlow Lite version:" -ForegroundColor Gray
Write-Host "   https://storage.googleapis.com/download.tensorflow.org/models/tflite/coco_ssd_mobilenet_v1_1.0_quant_2018_06_29.zip" -ForegroundColor White

Write-Host ""
Write-Host "📋 Downloaded models:" -ForegroundColor Green
if (Test-Path $modelsDir) {
    Get-ChildItem $modelsDir -Filter "*.onnx" | ForEach-Object {
        $sizeMB = [math]::Round($_.Length / 1MB, 2)
        Write-Host "  • $($_.Name) ($sizeMB MB)" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "🔧 To use in your app, update environment variable:" -ForegroundColor Cyan
Write-Host "   MODEL_PATH=./internal/services/AI/models/ssd_mobilenet_v2_lite.onnx" -ForegroundColor White
Write-Host "Or for ultra-fast processing:" -ForegroundColor Cyan
Write-Host "   MODEL_PATH=./internal/services/AI/models/yolov5n_lite.onnx" -ForegroundColor White

Write-Host ""
Write-Host "🎉 Download completed!" -ForegroundColor Green
