package scheduler

import (
	"time"

	"gorm.io/gorm"
)

// BackupType constants for the Type field
const (
	BackupTypeAutomated = "automated"
	BackupTypeManual    = "manual"
)

// BackupStatus constants for the Status field
const (
	BackupStatusCompleted  = "completed"
	BackupStatusFailed     = "failed"
	BackupStatusInProgress = "in_progress"
)

// ScheduleType constants for the Type field
const (
	ScheduleTypeBackup  = "backup"
	ScheduleTypeRestart = "restart"
	ScheduleTypeSave    = "save"
)

// ScheduleIntervalType constants for the ScheduleType field
const (
	ScheduleIntervalTypeInterval = "interval"
	ScheduleIntervalTypeTime     = "time"
)

// Backup stores metadata about each backup
type Backup struct {
	gorm.Model
	Filename     string `json:"filename" gorm:"not null"`
	OriginalSave string `json:"original_save" gorm:"not null"`
	Size         int64  `json:"size"`
	Checksum     string `json:"checksum"`               // SHA256
	Type         string `json:"type" gorm:"not null"`   // "automated" or "manual"
	Status       string `json:"status" gorm:"not null"` // "completed", "failed", "in_progress"
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedBy    string `json:"created_by,omitempty"` // username who triggered it

	// Computed fields for API compatibility (not stored in DB)
	BackupName string `json:"backup_name" gorm:"-"`
	SaveName   string `json:"save_name" gorm:"-"`
	BackupPath string `json:"backup_path" gorm:"-"`
}

// TableName returns the table name for the Backup model
func (Backup) TableName() string {
	return "backups"
}

// IsCompleted returns true if the backup completed successfully
func (b *Backup) IsCompleted() bool {
	return b.Status == BackupStatusCompleted
}

// IsFailed returns true if the backup failed
func (b *Backup) IsFailed() bool {
	return b.Status == BackupStatusFailed
}

// IsInProgress returns true if the backup is currently in progress
func (b *Backup) IsInProgress() bool {
	return b.Status == BackupStatusInProgress
}

// IsAutomated returns true if the backup was created by an automated schedule
func (b *Backup) IsAutomated() bool {
	return b.Type == BackupTypeAutomated
}

// IsManual returns true if the backup was created manually
func (b *Backup) IsManual() bool {
	return b.Type == BackupTypeManual
}

// Schedule stores scheduled task configuration
type Schedule struct {
	gorm.Model
	Name           string     `json:"name" gorm:"not null"`
	Type           string     `json:"type" gorm:"not null"` // "backup", "restart", "save"
	Enabled        bool       `json:"enabled" gorm:"default:true"`
	ScheduleType   string     `json:"schedule_type" gorm:"not null"` // "interval" or "time"
	IntervalMin    int        `json:"interval_min"`                  // For interval type (minutes)
	TimeOfDay      string     `json:"time_of_day"`                   // HH:MM for time type
	RetentionCount int        `json:"retention_count"`               // For backups - how many to keep
	WarningMinutes int        `json:"warning_minutes"`               // For restarts - warn players
	LastRun        *time.Time `json:"last_run,omitempty"`
	NextRun        *time.Time `json:"next_run,omitempty"`
}

// TableName returns the table name for the Schedule model
func (Schedule) TableName() string {
	return "schedules"
}

// IsEnabled returns true if the schedule is enabled
func (s *Schedule) IsEnabled() bool {
	return s.Enabled
}

// IsBackupSchedule returns true if this is a backup schedule
func (s *Schedule) IsBackupSchedule() bool {
	return s.Type == ScheduleTypeBackup
}

// IsRestartSchedule returns true if this is a restart schedule
func (s *Schedule) IsRestartSchedule() bool {
	return s.Type == ScheduleTypeRestart
}

// IsSaveSchedule returns true if this is a save schedule
func (s *Schedule) IsSaveSchedule() bool {
	return s.Type == ScheduleTypeSave
}

// IsIntervalBased returns true if this schedule runs at intervals
func (s *Schedule) IsIntervalBased() bool {
	return s.ScheduleType == ScheduleIntervalTypeInterval
}

// IsTimeBased returns true if this schedule runs at a specific time
func (s *Schedule) IsTimeBased() bool {
	return s.ScheduleType == ScheduleIntervalTypeTime
}

// CalculateNextRun calculates the next run time based on the schedule configuration
func (s *Schedule) CalculateNextRun(from time.Time) time.Time {
	if s.IsIntervalBased() {
		return from.Add(time.Duration(s.IntervalMin) * time.Minute)
	}

	// For time-based schedules, parse TimeOfDay and find next occurrence
	if s.TimeOfDay == "" {
		return from.Add(24 * time.Hour) // Default to 24 hours if no time specified
	}

	// Parse HH:MM format
	t, err := time.Parse("15:04", s.TimeOfDay)
	if err != nil {
		return from.Add(24 * time.Hour) // Default to 24 hours on parse error
	}

	// Create next run time with today's date
	nextRun := time.Date(
		from.Year(), from.Month(), from.Day(),
		t.Hour(), t.Minute(), 0, 0,
		from.Location(),
	)

	// If the time has already passed today, schedule for tomorrow
	if !nextRun.After(from) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	return nextRun
}

// UpdateNextRun updates the NextRun field based on current time
func (s *Schedule) UpdateNextRun() {
	now := time.Now()
	s.LastRun = &now
	nextRun := s.CalculateNextRun(now)
	s.NextRun = &nextRun
}

// ShouldRun returns true if the schedule should run now
func (s *Schedule) ShouldRun() bool {
	if !s.Enabled {
		return false
	}

	if s.NextRun == nil {
		return true
	}

	return time.Now().After(*s.NextRun) || time.Now().Equal(*s.NextRun)
}
