package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/OpenFactorioServerManager/factorio-server-manager/bootstrap"
	"github.com/OpenFactorioServerManager/factorio-server-manager/factorio"
	"gorm.io/gorm"
)

// GetBackupDir returns the backup directory path
func GetBackupDir() string {
	config := bootstrap.GetConfig()
	return filepath.Join(config.FactorioDir, "backups")
}

// ensureBackupDir creates the backup directory if it doesn't exist
func ensureBackupDir() error {
	backupDir := GetBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	return nil
}

// calculateChecksum computes SHA256 checksum of a file
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// Set same permissions as source
	if err := os.Chmod(dst, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// triggerServerSave sends RCON command to trigger a server save
func triggerServerSave() error {
	server := factorio.GetFactorioServer()
	if server == nil {
		return errors.New("factorio server not initialized")
	}

	if !server.GetRunning() {
		return errors.New("server is not running")
	}

	rc := server.GetRcon()
	if rc == nil {
		return errors.New("RCON not connected")
	}

	// Send the /server-save command
	reqID, err := rc.Write("/server-save")
	if err != nil {
		return fmt.Errorf("failed to send server-save command: %w", err)
	}

	log.Printf("Sent /server-save command with request ID: %d", reqID)

	// Read the response to confirm
	resp, _, err := rc.Read()
	if err != nil {
		log.Printf("Warning: failed to read RCON response: %v", err)
		// Don't fail the backup, the save might still succeed
	} else {
		log.Printf("RCON response: %s", resp)
	}

	// Wait a moment for the save to complete
	// Factorio saves are typically fast, but we give it a small buffer
	time.Sleep(2 * time.Second)

	return nil
}

// getBackupFilePath returns the full path to a backup file
func getBackupFilePath(filename string) string {
	return filepath.Join(GetBackupDir(), filename)
}

// CreateBackup creates a backup of the specified save file
// If the server is running, it triggers a server save first via RCON
// backupType should be BackupTypeManual or BackupTypeAutomated
func CreateBackup(db *gorm.DB, saveName string, backupType string, createdBy string) (*Backup, error) {
	if saveName == "" {
		return nil, errors.New("save name cannot be empty")
	}

	config := bootstrap.GetConfig()
	server := factorio.GetFactorioServer()

	// Ensure backup directory exists
	if err := ensureBackupDir(); err != nil {
		return nil, err
	}

	// Check if server is running
	serverRunning := false
	if server != nil && server.GetRunning() {
		serverRunning = true
	}

	// Create backup record with in_progress status
	timestamp := time.Now().UTC()
	backupFilename := fmt.Sprintf("%s_%s.zip",
		saveName[:len(saveName)-len(filepath.Ext(saveName))],
		timestamp.Format("20060102_150405"))

	backup := &Backup{
		Filename:     backupFilename,
		OriginalSave: saveName,
		Type:         backupType,
		Status:       BackupStatusInProgress,
		CreatedBy:    createdBy,
	}

	// Save initial record to database
	if err := db.Create(backup).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}

	// If server is running, trigger a save first
	if serverRunning {
		log.Printf("Server is running, triggering /server-save before backup")
		if err := triggerServerSave(); err != nil {
			log.Printf("Warning: failed to trigger server save: %v", err)
			// Continue with backup anyway - the existing save file should still be valid
		}
	}

	// Build source path
	sourcePath := filepath.Join(config.FactorioSavesDir, saveName)

	// Verify source file exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		backup.Status = BackupStatusFailed
		if os.IsNotExist(err) {
			backup.ErrorMessage = fmt.Sprintf("save file not found: %s", saveName)
		} else {
			backup.ErrorMessage = fmt.Sprintf("failed to stat source file: %v", err)
		}
		db.Save(backup)
		return nil, errors.New(backup.ErrorMessage)
	}

	backupPath := getBackupFilePath(backupFilename)

	// Copy the save file to backup location
	if err := copyFile(sourcePath, backupPath); err != nil {
		backup.Status = BackupStatusFailed
		backup.ErrorMessage = fmt.Sprintf("failed to copy save file: %v", err)
		db.Save(backup)
		return nil, errors.New(backup.ErrorMessage)
	}

	// Calculate checksum of the backup
	checksum, err := calculateChecksum(backupPath)
	if err != nil {
		// Clean up the backup file if checksum fails
		os.Remove(backupPath)
		backup.Status = BackupStatusFailed
		backup.ErrorMessage = fmt.Sprintf("failed to calculate backup checksum: %v", err)
		db.Save(backup)
		return nil, errors.New(backup.ErrorMessage)
	}

	// Update backup record with success
	backup.Size = sourceInfo.Size()
	backup.Checksum = checksum
	backup.Status = BackupStatusCompleted

	if err := db.Save(backup).Error; err != nil {
		// Clean up the backup file if DB save fails
		os.Remove(backupPath)
		return nil, fmt.Errorf("failed to update backup record: %w", err)
	}

	// Populate computed fields for API compatibility
	populateComputedFields(backup)

	log.Printf("Created backup: %s (checksum: %s)", backupFilename, checksum)
	return backup, nil
}

// RestoreBackup restores a backup to the saves directory
// The server must be stopped before restoring
func RestoreBackup(db *gorm.DB, backupID uint) error {
	// Check if server is running
	server := factorio.GetFactorioServer()
	if server != nil && server.GetRunning() {
		return errors.New("cannot restore backup while server is running - please stop the server first")
	}

	// Find the backup record
	var backup Backup
	if err := db.First(&backup, backupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("backup not found with ID: %d", backupID)
		}
		return fmt.Errorf("failed to find backup: %w", err)
	}

	// Check backup status
	if backup.IsFailed() {
		return fmt.Errorf("cannot restore failed backup: %s", backup.ErrorMessage)
	}
	if backup.IsInProgress() {
		return errors.New("cannot restore backup that is still in progress")
	}

	backupPath := getBackupFilePath(backup.Filename)

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found at path: %s", backupPath)
	}

	// Verify checksum before restore
	valid, err := VerifyBackup(db, backupID)
	if err != nil {
		return fmt.Errorf("failed to verify backup: %w", err)
	}
	if !valid {
		return errors.New("backup checksum verification failed - backup may be corrupted")
	}

	config := bootstrap.GetConfig()
	destPath := filepath.Join(config.FactorioSavesDir, backup.OriginalSave)

	// Create backup of existing save file if it exists (safety measure)
	if _, err := os.Stat(destPath); err == nil {
		preRestoreBackup := destPath + ".pre-restore"
		if err := copyFile(destPath, preRestoreBackup); err != nil {
			log.Printf("Warning: failed to create pre-restore backup: %v", err)
		}
	}

	// Copy backup to saves directory
	if err := copyFile(backupPath, destPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	log.Printf("Restored backup %s to %s", backup.Filename, destPath)
	return nil
}

// VerifyBackup verifies the checksum of a backup file
func VerifyBackup(db *gorm.DB, backupID uint) (bool, error) {
	// Find the backup record
	var backup Backup
	if err := db.First(&backup, backupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("backup not found with ID: %d", backupID)
		}
		return false, fmt.Errorf("failed to find backup: %w", err)
	}

	// Failed or in-progress backups cannot be verified
	if backup.IsFailed() || backup.IsInProgress() {
		return false, errors.New("backup is not in a completed state")
	}

	backupPath := getBackupFilePath(backup.Filename)

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return false, fmt.Errorf("backup file not found at path: %s", backupPath)
	}

	// Calculate current checksum
	currentChecksum, err := calculateChecksum(backupPath)
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Compare with stored checksum
	if currentChecksum != backup.Checksum {
		log.Printf("Checksum mismatch for backup %s: expected %s, got %s",
			backup.Filename, backup.Checksum, currentChecksum)
		return false, nil
	}

	log.Printf("Backup %s verified successfully", backup.Filename)
	return true, nil
}

// DeleteBackup deletes a backup file and its database record
func DeleteBackup(db *gorm.DB, backupID uint) error {
	// Find the backup record
	var backup Backup
	if err := db.First(&backup, backupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("backup not found with ID: %d", backupID)
		}
		return fmt.Errorf("failed to find backup: %w", err)
	}

	backupPath := getBackupFilePath(backup.Filename)

	// Delete the backup file if it exists
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			return fmt.Errorf("failed to delete backup file: %w", err)
		}
		log.Printf("Deleted backup file: %s", backupPath)
	}

	// Delete the database record (hard delete)
	if err := db.Unscoped().Delete(&backup).Error; err != nil {
		return fmt.Errorf("failed to delete backup record: %w", err)
	}

	log.Printf("Deleted backup record: %s (ID: %d)", backup.Filename, backupID)
	return nil
}

// CleanupOldBackups enforces retention policy by deleting old backups
// when the count exceeds the retention limit
// Keeps the most recent backups and deletes the oldest ones
func CleanupOldBackups(db *gorm.DB, retentionCount int) error {
	if retentionCount <= 0 {
		return errors.New("retention count must be greater than 0")
	}

	// Get all completed backups ordered by creation date (newest first)
	var backups []Backup
	if err := db.Where("status = ?", BackupStatusCompleted).Order("created_at DESC").Find(&backups).Error; err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	totalBackups := len(backups)
	if totalBackups <= retentionCount {
		log.Printf("Backup count (%d) within retention limit (%d), no cleanup needed",
			totalBackups, retentionCount)
		return nil
	}

	// Sort by creation date to ensure we delete the oldest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	// Delete backups beyond the retention count
	deletedCount := 0
	for i := retentionCount; i < len(backups); i++ {
		backup := backups[i]
		if err := DeleteBackup(db, backup.ID); err != nil {
			log.Printf("Warning: failed to delete backup %s: %v", backup.Filename, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Cleanup complete: deleted %d backups, kept %d", deletedCount, retentionCount)
	return nil
}

// CleanupOldBackupsForSave enforces retention policy for a specific save file
func CleanupOldBackupsForSave(db *gorm.DB, saveName string, retentionCount int) error {
	if retentionCount <= 0 {
		return errors.New("retention count must be greater than 0")
	}

	if saveName == "" {
		return errors.New("save name cannot be empty")
	}

	// Get completed backups for this save, ordered by creation date (newest first)
	var backups []Backup
	if err := db.Where("original_save = ? AND status = ?", saveName, BackupStatusCompleted).
		Order("created_at DESC").Find(&backups).Error; err != nil {
		return fmt.Errorf("failed to list backups for save %s: %w", saveName, err)
	}

	totalBackups := len(backups)
	if totalBackups <= retentionCount {
		log.Printf("Backup count for %s (%d) within retention limit (%d), no cleanup needed",
			saveName, totalBackups, retentionCount)
		return nil
	}

	// Delete backups beyond the retention count
	deletedCount := 0
	for i := retentionCount; i < len(backups); i++ {
		backup := backups[i]
		if err := DeleteBackup(db, backup.ID); err != nil {
			log.Printf("Warning: failed to delete backup %s: %v", backup.Filename, err)
			continue
		}
		deletedCount++
	}

	log.Printf("Cleanup for %s complete: deleted %d backups, kept %d", saveName, deletedCount, retentionCount)
	return nil
}

// ListBackups returns all backups, optionally filtered by save name
func ListBackups(db *gorm.DB, saveName string) ([]Backup, error) {
	var backups []Backup
	query := db.Order("created_at DESC")

	if saveName != "" {
		query = query.Where("original_save = ?", saveName)
	}

	if err := query.Find(&backups).Error; err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// Populate computed fields for API compatibility
	populateComputedFieldsSlice(backups)

	return backups, nil
}

// GetBackup returns a backup by ID
func GetBackup(db *gorm.DB, backupID uint) (*Backup, error) {
	var backup Backup
	if err := db.First(&backup, backupID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("backup not found with ID: %d", backupID)
		}
		return nil, fmt.Errorf("failed to find backup: %w", err)
	}
	// Populate computed fields for API compatibility
	populateComputedFields(&backup)
	return &backup, nil
}

// BackupStats holds statistics about backups
type BackupStats struct {
	TotalCount      int64     `json:"total_count"`
	CompletedCount  int64     `json:"completed_count"`
	FailedCount     int64     `json:"failed_count"`
	TotalSize       int64     `json:"total_size"`
	OldestBackup    time.Time `json:"oldest_backup,omitempty"`
	NewestBackup    time.Time `json:"newest_backup,omitempty"`
	UniqueSaveCount int64     `json:"unique_save_count"`
}

// GetBackupStats returns statistics about all backups
func GetBackupStats(db *gorm.DB) (*BackupStats, error) {
	stats := &BackupStats{}

	// Total count
	if err := db.Model(&Backup{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count backups: %w", err)
	}

	if stats.TotalCount == 0 {
		return stats, nil
	}

	// Completed count
	if err := db.Model(&Backup{}).Where("status = ?", BackupStatusCompleted).Count(&stats.CompletedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed backups: %w", err)
	}

	// Failed count
	if err := db.Model(&Backup{}).Where("status = ?", BackupStatusFailed).Count(&stats.FailedCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed backups: %w", err)
	}

	// Total size (only for completed backups)
	if err := db.Model(&Backup{}).Where("status = ?", BackupStatusCompleted).
		Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize).Error; err != nil {
		return nil, fmt.Errorf("failed to sum backup sizes: %w", err)
	}

	// Oldest completed backup
	var oldest Backup
	if err := db.Where("status = ?", BackupStatusCompleted).Order("created_at ASC").First(&oldest).Error; err == nil {
		stats.OldestBackup = oldest.CreatedAt
	}

	// Newest completed backup
	var newest Backup
	if err := db.Where("status = ?", BackupStatusCompleted).Order("created_at DESC").First(&newest).Error; err == nil {
		stats.NewestBackup = newest.CreatedAt
	}

	// Distinct save file count
	if err := db.Model(&Backup{}).Distinct("original_save").Count(&stats.UniqueSaveCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count distinct saves: %w", err)
	}

	return stats, nil
}

// InitBackupSystem ensures the backup directory and database table exist
func InitBackupSystem(db *gorm.DB) error {
	// Create backup directory
	if err := ensureBackupDir(); err != nil {
		return err
	}

	// AutoMigrate the Backup model
	if err := db.AutoMigrate(&Backup{}); err != nil {
		return fmt.Errorf("failed to migrate backup table: %w", err)
	}

	log.Println("Backup system initialized")
	return nil
}

// InitBackupDatabase is an alias for InitBackupSystem for backward compatibility
func InitBackupDatabase(db *gorm.DB) error {
	return InitBackupSystem(db)
}

// populateComputedFields sets the computed fields on a backup struct
func populateComputedFields(backup *Backup) {
	backup.BackupName = backup.Filename
	backup.SaveName = backup.OriginalSave
	backup.BackupPath = getBackupFilePath(backup.Filename)
}

// populateComputedFieldsSlice sets computed fields on a slice of backups
func populateComputedFieldsSlice(backups []Backup) {
	for i := range backups {
		populateComputedFields(&backups[i])
	}
}
