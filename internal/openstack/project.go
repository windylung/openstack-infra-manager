package openstack

import (
	"context"
	"fmt"
	"strings"

	"example.com/quotaapi/internal/models"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/users"
)

// ProjectManager handles OpenStack project operations for students
type ProjectManager struct {
	clients  *Clients
	domainID string // 캐시
}

// NewProjectManager creates a new project manager
func NewProjectManager(clients *Clients) *ProjectManager {
	return &ProjectManager{clients: clients}
}

// ensureDomainID gets and caches the default domain ID
func (pm *ProjectManager) ensureDomainID(ctx context.Context) (string, error) {
	if pm.domainID != "" {
		return pm.domainID, nil
	}

	// Domain 이름이 "Default"라는 가정 (환경에 맞게 변경 가능)
	pages, err := domains.List(pm.clients.Identity, domains.ListOpts{Name: "Default"}).AllPages(ctx)
	if err != nil {
		return "", fmt.Errorf("list domains: %w", err)
	}

	ds, err := domains.ExtractDomains(pages)
	if err != nil || len(ds) == 0 {
		return "", fmt.Errorf("domain 'Default' not found")
	}

	pm.domainID = ds[0].ID
	return pm.domainID, nil
}

// ptrBool helper function
func ptrBool(b bool) *bool {
	return &b
}

// CreateStudentProject creates a complete student environment in OpenStack
func (pm *ProjectManager) CreateStudentProject(ctx context.Context, student *models.Student) error {
	// 1. 프로젝트 생성
	project, err := pm.createProject(ctx, student)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// 2. 사용자 생성
	user, err := pm.createUser(ctx, student)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 3. 사용자를 프로젝트에 할당하고 기본 역할 부여
	if err := pm.assignUserToProjectWithRoles(ctx, user.ID, project.ID); err != nil {
		return fmt.Errorf("failed to assign user to project: %w", err)
	}

	// 4. 기본 쿼타 설정 (basic 프로파일)
	if err := pm.setDefaultQuotas(ctx, project.ID); err != nil {
		return fmt.Errorf("failed to set default quotas: %w", err)
	}

	// 5. 학생 정보 업데이트
	student.KeystoneProjectID = project.ID
	student.KeystoneUserID = user.ID

	return nil
}

// createProject creates a new project for a student
func (pm *ProjectManager) createProject(ctx context.Context, student *models.Student) (*projects.Project, error) {
	did, err := pm.ensureDomainID(ctx)
	if err != nil {
		return nil, err
	}

	projectName := fmt.Sprintf("student-%s-project", student.StudentID)
	projectDescription := fmt.Sprintf("Project for student %s (%s)", student.Name, student.StudentID)

	createOpts := projects.CreateOpts{
		Name:        projectName,
		Description: projectDescription,
		DomainID:    did, // 올바른 DomainID 사용
		Enabled:     ptrBool(true),
	}

	project, err := projects.Create(ctx, pm.clients.Identity, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("Created project: %s (ID: %s)\n", projectName, project.ID)
	return project, nil
}

// createUser creates a new user for a student
func (pm *ProjectManager) createUser(ctx context.Context, student *models.Student) (*users.User, error) {
	did, err := pm.ensureDomainID(ctx)
	if err != nil {
		return nil, err
	}

	userName := fmt.Sprintf("student-%s-user", student.StudentID)
	userDescription := fmt.Sprintf("User account for student %s (%s)", student.Name, student.StudentID)

	// 임시 비밀번호 생성 (실제로는 안전한 방법 사용)
	tempPassword := fmt.Sprintf("temp-%s-2025", student.StudentID)

	createOpts := users.CreateOpts{
		Name:        userName,
		Description: userDescription,
		Password:    tempPassword,
		DomainID:    did, // 올바른 DomainID 사용
		Enabled:     ptrBool(true),
	}

	user, err := users.Create(ctx, pm.clients.Identity, createOpts).Extract()
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("Created user: %s (ID: %s)\n", userName, user.ID)
	return user, nil
}

// assignUserToProjectWithRoles assigns a user to a project with default roles
func (pm *ProjectManager) assignUserToProjectWithRoles(ctx context.Context, userID, projectID string) error {
	// 기본 역할: member (실제 존재하는 롤)
	role, err := pm.findRoleByName(ctx, "member")
	if err != nil {
		return fmt.Errorf("member role not found: %w", err)
	}

	// 핵심: 실제 롤 부여
	if err := roles.Assign(ctx, pm.clients.Identity, role.ID, roles.AssignOpts{
		UserID:    userID,
		ProjectID: projectID,
	}).ExtractErr(); err != nil {
		return fmt.Errorf("assign role to user on project: %w", err)
	}

	fmt.Printf("Assigned role %s to user %s on project %s\n", role.Name, userID, projectID)
	return nil
}

// findRoleByName finds a role by name
func (pm *ProjectManager) findRoleByName(ctx context.Context, roleName string) (*roles.Role, error) {
	// 역할 목록 조회
	allPages, err := roles.List(pm.clients.Identity, roles.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roleList, err := roles.ExtractRoles(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract roles: %w", err)
	}

	// 이름으로 역할 찾기
	for _, role := range roleList {
		if strings.EqualFold(role.Name, roleName) {
			return &role, nil
		}
	}

	return nil, fmt.Errorf("role %s not found", roleName)
}

func (pm *ProjectManager) DeleteStudentProject(ctx context.Context, student *models.Student) error {
	if student.KeystoneProjectID == "" {
		return fmt.Errorf("no project ID found for student %s", student.StudentID)
	}

	// 1. 프로젝트 삭제 (사용자 할당도 자동으로 제거됨)
	if err := projects.Delete(ctx, pm.clients.Identity, student.KeystoneProjectID).ExtractErr(); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// 2. 사용자 삭제
	if student.KeystoneUserID != "" {
		if err := users.Delete(ctx, pm.clients.Identity, student.KeystoneUserID).ExtractErr(); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}

	fmt.Printf("Deleted project and user for student %s\n", student.StudentID)
	return nil
}

// ListAllProjects lists all projects in OpenStack
func (pm *ProjectManager) ListAllProjects(ctx context.Context) ([]projects.Project, error) {
	// 프로젝트 목록 조회
	allPages, err := projects.List(pm.clients.Identity, projects.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projectList, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, fmt.Errorf("failed to extract projects: %w", err)
	}

	return projectList, nil
}

// FindStudentProject finds a project by student ID pattern
func (pm *ProjectManager) FindStudentProject(ctx context.Context, studentID string) (*projects.Project, error) {
	allProjects, err := pm.ListAllProjects(ctx)
	if err != nil {
		return nil, err
	}

	// student-{studentID}-project 패턴으로 검색
	expectedName := fmt.Sprintf("student-%s-project", studentID)

	for _, project := range allProjects {
		if project.Name == expectedName {
			return &project, nil
		}
	}

	return nil, fmt.Errorf("project not found for student %s", studentID)
}

// setDefaultQuotas sets default quotas for a student project using basic profile
func (pm *ProjectManager) setDefaultQuotas(ctx context.Context, projectID string) error {
	// models.Profiles의 basic 프로파일 사용
	basicProfile := models.GetBasicProfile()

	// Nova 쿼타 설정 (basic 프로파일 값)
	cores := basicProfile.Cores
	ramMB := basicProfile.RAMMB
	instances := basicProfile.Instances
	if err := pm.clients.ApplyNovaQuota(ctx, projectID, &cores, &ramMB, &instances); err != nil {
		return fmt.Errorf("failed to set Nova quotas: %w", err)
	}

	// Cinder 쿼타 설정 (basic 프로파일 값)
	volumes := basicProfile.Volumes
	gigabytes := basicProfile.Gigabytes
	snapshots := basicProfile.Snapshots
	if err := pm.clients.ApplyCinderQuota(ctx, projectID, &volumes, &snapshots, &gigabytes); err != nil {
		return fmt.Errorf("failed to set Cinder quotas: %w", err)
	}

	// Neutron 쿼타 설정
	ports := basicProfile.Ports
	floatingIPs := basicProfile.FloatingIPs
	if err := pm.clients.ApplyNeutronQuota(ctx, projectID, &ports, &floatingIPs); err != nil {
		return fmt.Errorf("failed to set Neutron quotas: %w", err)
	}

	fmt.Printf("Set basic profile quotas for project %s: vCPU=%d, RAM=%dMB, Instances=%d, Volumes=%d, Disk=%dGB, Ports=%d, FloatingIPs=%d\n",
		projectID, cores, ramMB, instances, volumes, gigabytes, ports, floatingIPs)
	return nil
}

// GetClients returns the OpenStack clients
func (pm *ProjectManager) GetClients() *Clients {
	return pm.clients
}
