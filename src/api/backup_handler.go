package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/OpenFactorioServerManager/factorio-server-manager/scheduler"
	"github.com/gorilla/mux"
)

// ListBackups - GET /api/backups/list
// Returns list of all backups with pagination support
func ListBackups(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Optional filter by save name
	saveName := r.URL.Query().Get("save_name")

	var backups []scheduler.Backup
	var total int64

	// Build query
	query := GetDB().Model(&scheduler.Backup{})
	if saveName != "" {
		query = query.Where("save_name = ?", saveName)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		resp = fmt.Sprintf("Error counting backups: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get paginated results, ordered by backed_up_at time descending
	offset := (page - 1) * pageSize
	if err := query.Order("backed_up_at DESC").Offset(offset).Limit(pageSize).Find(&backups).Error; err != nil {
		resp = fmt.Sprintf("Error listing backups: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp = map[string]interface{}{
		"backups":   backups,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
}

// CreateBackupHandler - POST /api/backups/create
// Body: {"save_name": "mysave.zip"}
// Creates a manual backup of the specified save
func CreateBackupHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	var data struct {
		SaveName string `json:"save_name"`
	}

	resp, err := ReadFromRequestBody(w, r, &data)
	if err != nil {
		return
	}

	if data.SaveName == "" {
		resp = "save_name is required"
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate save path to prevent directory traversal
	_, err = ValidatePathComponent(data.SaveName)
	if err != nil {
		resp = fmt.Sprintf("Invalid save name: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get current user from session for created_by field
	createdBy := ""
	session, _, sessionErr := ReadSessionStore(w, r, "authentication")
	if sessionErr == nil {
		if usernameVal, ok := session.Values["username"]; ok {
			if username, ok := usernameVal.(string); ok {
				createdBy = username
			}
		}
	}

	// Create backup using scheduler package function
	backup, err := scheduler.CreateBackup(GetDB(), data.SaveName, scheduler.BackupTypeManual, createdBy)
	if err != nil {
		resp = fmt.Sprintf("Error creating backup: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Created backup: %s (size: %d, checksum: %s)", backup.BackupName, backup.Size, backup.Checksum)
	resp = backup
}

// RestoreBackupHandler - POST /api/backups/restore/{id}
// Restores the specified backup (requires server to be stopped)
func RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp = fmt.Sprintf("Invalid backup ID: %s", idStr)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get backup record first to return useful information
	backup, err := scheduler.GetBackup(GetDB(), uint(id))
	if err != nil {
		resp = fmt.Sprintf("Backup not found: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Restore backup using scheduler package function (checks server status internally)
	err = scheduler.RestoreBackup(GetDB(), uint(id))
	if err != nil {
		// Check if the error is because server is running
		if err.Error() == "cannot restore backup while server is running - please stop the server first" {
			resp = err.Error()
			log.Println(resp)
			w.WriteHeader(http.StatusConflict)
			return
		}
		resp = fmt.Sprintf("Error restoring backup: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Restored backup %d (%s) to %s", backup.ID, backup.BackupName, backup.SaveName)
	resp = map[string]interface{}{
		"message":   fmt.Sprintf("Backup restored successfully to %s", backup.SaveName),
		"backup_id": backup.ID,
		"save_name": backup.SaveName,
	}
}

// DeleteBackupHandler - DELETE /api/backups/{id}
// Deletes a backup
func DeleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp = fmt.Sprintf("Invalid backup ID: %s", idStr)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get backup record first to check if it exists and for logging
	backup, err := scheduler.GetBackup(GetDB(), uint(id))
	if err != nil {
		resp = fmt.Sprintf("Backup not found: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	backupName := backup.BackupName

	// Delete backup using scheduler package function
	err = scheduler.DeleteBackup(GetDB(), uint(id))
	if err != nil {
		resp = fmt.Sprintf("Error deleting backup: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Deleted backup: %d (%s)", id, backupName)
	resp = fmt.Sprintf("Backup %d deleted successfully", id)
}

// DownloadBackupHandler - GET /api/backups/download/{id}
// Downloads the backup file
func DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Printf("Invalid backup ID: %s", idStr)
		http.Error(w, "Invalid backup ID", http.StatusBadRequest)
		return
	}

	// Get backup record
	backup, err := scheduler.GetBackup(GetDB(), uint(id))
	if err != nil {
		log.Printf("Backup not found: %s", err)
		http.Error(w, "Backup not found", http.StatusNotFound)
		return
	}

	// Check if file exists
	if _, err := os.Stat(backup.BackupPath); os.IsNotExist(err) {
		log.Printf("Backup file not found on disk: %s", backup.BackupPath)
		http.Error(w, "Backup file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", backup.BackupName))
	log.Printf("%s downloading backup: %s", r.Host, backup.BackupName)

	http.ServeFile(w, r, backup.BackupPath)
}

// VerifyBackupHandler - GET /api/backups/verify/{id}
// Verifies the backup checksum
func VerifyBackupHandler(w http.ResponseWriter, r *http.Request) {
	var resp interface{}
	defer func() {
		WriteResponse(w, resp)
	}()

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp = fmt.Sprintf("Invalid backup ID: %s", idStr)
		log.Println(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get backup record for response info
	backup, err := scheduler.GetBackup(GetDB(), uint(id))
	if err != nil {
		resp = fmt.Sprintf("Backup not found: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Verify backup using scheduler package function
	isValid, err := scheduler.VerifyBackup(GetDB(), uint(id))
	if err != nil {
		// Check if it's a file not found error
		if os.IsNotExist(err) {
			resp = map[string]interface{}{
				"valid":             false,
				"error":             "Backup file not found on disk",
				"backup_id":         backup.ID,
				"expected_checksum": backup.Checksum,
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp = fmt.Sprintf("Error verifying backup: %s", err)
		log.Println(resp)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Verified backup %d: valid=%t", backup.ID, isValid)

	resp = map[string]interface{}{
		"valid":             isValid,
		"backup_id":         backup.ID,
		"expected_checksum": backup.Checksum,
	}

	if !isValid {
		w.WriteHeader(http.StatusConflict)
	}
}
