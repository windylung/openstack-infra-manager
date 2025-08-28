package models

import (
	"time"
)

// Course represents a course in the system
type Course struct {
	CourseID     string          `json:"course_id"`
	Title        string          `json:"title"`
	Department   string          `json:"department"`
	Semester     string          `json:"semester"`
	StartAt      time.Time       `json:"start_at"`
	EndAt        time.Time       `json:"end_at"`
	QuotaProfile QuotaProfile    `json:"quota_profile"`
	Defaults     *CourseDefaults `json:"defaults,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// QuotaProfile represents the quota settings for a course
type QuotaProfile struct {
	Instances   int `json:"instances"`
	Cores       int `json:"cores"`
	RAMMB       int `json:"ramMB"`
	Volumes     int `json:"volumes"`
	Gigabytes   int `json:"gigabytes"`
	Ports       int `json:"ports"`
	FloatingIPs int `json:"floatingIPs"`
	Snapshots   int `json:"snapshots"`
}

// Profiles: 사전에 정의된 프로파일 목록
// - 사용 예: profile=basic | lab
// - 필요한 경우 향후 설정 파일/DB로 분리 가능
var Profiles = map[string]QuotaProfile{
	"basic": {Cores: 8, RAMMB: 16384, Instances: 10, Gigabytes: 100, Volumes: 10, Snapshots: 10, Ports: 10, FloatingIPs: 5},
	"lab":   {Cores: 16, RAMMB: 32768, Instances: 20, Gigabytes: 200, Volumes: 20, Snapshots: 20, Ports: 20, FloatingIPs: 10},
}

// GetBasicProfile returns the basic quota profile
func GetBasicProfile() QuotaProfile {
	return Profiles["basic"]
}

// GetLabProfile returns the lab quota profile
func GetLabProfile() QuotaProfile {
	return Profiles["lab"]
}

// CourseDefaults - 최소 필드만 유지
type CourseDefaults struct {
	ImageID           string   `json:"imageId,omitempty"`
	FlavorIDs         []string `json:"flavorIds,omitempty"`
	NetworkID         string   `json:"networkId,omitempty"`
	ExternalNetworkID string   `json:"externalNetworkId,omitempty"`
	SecurityGroup     string   `json:"securityGroup,omitempty"`
	BootFromVolume    bool     `json:"bootFromVolume,omitempty"`
	RootVolumeGB      int      `json:"rootVolumeGB,omitempty"`
}

// CourseCreateRequest represents the request to create a new course
type CourseCreateRequest struct {
	CourseID     string          `json:"course_id" validate:"required"`
	Title        string          `json:"title" validate:"required"`
	Department   string          `json:"department" validate:"required"`
	Semester     string          `json:"semester" validate:"required"`
	StartAt      string          `json:"start_at" validate:"required"`
	EndAt        string          `json:"end_at" validate:"required"`
	QuotaProfile QuotaProfile    `json:"quota_profile" validate:"required"`
	Defaults     *CourseDefaults `json:"defaults,omitempty"`
}

// CourseUpdateRequest represents the request to update a course
type CourseUpdateRequest struct {
	Title        *string         `json:"title,omitempty"`
	Department   *string         `json:"department,omitempty"`
	Semester     *string         `json:"semester,omitempty"`
	StartAt      *string         `json:"start_at,omitempty"`
	EndAt        *string         `json:"end_at,omitempty"`
	QuotaProfile *QuotaProfile   `json:"quota_profile,omitempty"`
	Defaults     *CourseDefaults `json:"defaults,omitempty"`
}
