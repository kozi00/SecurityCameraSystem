package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"webserver/internal/database"
)

func main() {
	imagesDir := flag.String("images", "static/images", "Directory containing images")
	dbPath := flag.String("db", "data/images.db", "Database path")
	flag.Parse()

	fmt.Printf("Migrating images from %s to database %s\n", *imagesDir, *dbPath)

	// Ensure database directory exists
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	// Initialize database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Scan images directory
	files, err := os.ReadDir(*imagesDir)
	if err != nil {
		log.Fatalf("Failed to read images directory: %v", err)
	}

	var images []database.Image
	skipped := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".jpg" {
			continue
		}

		timestamp, camera, objects, err := database.ParseFilename(file.Name())
		if err != nil {
			log.Printf("⚠️  Skipping %s: %v", file.Name(), err)
			skipped++
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Printf("⚠️  Failed to get info for %s: %v", file.Name(), err)
			skipped++
			continue
		}

		images = append(images, database.Image{
			Filename:  file.Name(),
			Camera:    camera,
			Objects:   objects,
			Timestamp: timestamp,
			FilePath:  filepath.Join(*imagesDir, file.Name()),
			FileSize:  info.Size(),
		})
	}

	if len(images) == 0 {
		fmt.Println("No images found to migrate")
		return
	}

	// Bulk insert
	fmt.Printf("Inserting %d images into database...\n", len(images))
	if err := db.BulkInsertImages(images); err != nil {
		log.Fatalf("Failed to insert images: %v", err)
	}

	fmt.Printf("✅ Successfully migrated %d images to database\n", len(images))
	if skipped > 0 {
		fmt.Printf("⚠️  Skipped %d files (invalid format or errors)\n", skipped)
	}

	// Show stats
	stats, err := db.GetStats()
	if err == nil {
		fmt.Printf("\n📊 Database Statistics:\n")
		fmt.Printf("   Total images: %v\n", stats["total_images"])
		fmt.Printf("   Total size: %v bytes\n", stats["total_size_bytes"])
		if perCamera, ok := stats["per_camera"].(map[string]int); ok {
			fmt.Printf("   Per camera:\n")
			for camera, count := range perCamera {
				fmt.Printf("      - %s: %d images\n", camera, count)
			}
		}
	}
}
