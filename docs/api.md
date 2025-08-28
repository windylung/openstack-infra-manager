# 🔗 API 명세서

## 📋 개요

Student Quota Management System의 REST API 명세서입니다.

## 🌐 기본 정보

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json`
- **인증**: OpenStack Keystone 토큰 기반

## 📊 응답 형식

### 성공 응답
```json
{
  "status": "success",
  "data": { ... },
  "message": "Operation completed successfully"
}
```

### 에러 응답
```json
{
  "status": "error",
  "error": "Error description",
  "code": "ERROR_CODE"
}
```

## 🔗 API 엔드포인트

### 1. 학생 관리 (Student Management)

#### 1.1 학생 목록 조회
```http
GET /students
```

**응답 예시:**
```json
[
  {
    "student_id": "32210003",
    "name": "이지수",
    "email": "jisoo@dankook.ac.kr",
    "department": "소프트웨어학과",
    "keystone_project_id": "98077908079746cab05e896ebf54258f",
    "keystone_user_id": "ef805f65998d4da7bff74d80c9b081fb",
    "created_at": "2025-08-28T02:26:03.589407Z"
  }
]
```

#### 1.2 학생 등록
```http
POST /students
```

**요청 본문:**
```json
{
  "student_id": "TEST001",
  "name": "테스트학생",
  "email": "test@university.ac.kr",
  "department": "테스트학과"
}
```

**응답 예시:**
```json
{
  "student_id": "TEST001",
  "name": "테스트학생",
  "email": "test@university.ac.kr",
  "department": "테스트학과",
  "keystone_project_id": "TEST001",
  "created_at": "2025-08-28T01:45:29.788108Z"
}
```

#### 1.3 학생 상세 조회
```http
GET /students/{student_id}
```

**응답 예시:**
```json
{
  "student_id": "32210003",
  "name": "이지수",
  "email": "jisoo@dankook.ac.kr",
  "department": "소프트웨어학과",
  "keystone_project_id": "98077908079746cab05e896ebf54258f",
  "keystone_user_id": "ef805f65998d4da7bff74d80c9b081fb",
  "created_at": "2025-08-28T02:26:03.589407Z"
}
```

### 2. 과목 관리 (Course Management)

#### 2.1 과목 목록 조회
```http
GET /courses
```

**응답 예시:**
```json
[
  {
    "course_id": "AI-2025-1",
    "title": "인공지능 기초",
    "department": "컴퓨터공학과",
    "semester": "2025-1",
    "start_at": "2025-03-01T00:00:00Z",
    "end_at": "2025-06-30T00:00:00Z",
    "quota_profile": {
      "instances": 2,
      "cores": 4,
      "ramMB": 8192,
      "volumes": 2,
      "gigabytes": 80,
      "ports": 2,
      "floatingIPs": 1,
      "snapshots": 0
    },
    "defaults": {
      "imageId": "ubuntu-22.04",
      "flavorIds": ["m1.medium"],
      "networkId": "private-network",
      "externalNetworkId": "public-network",
      "securityGroup": "ai-sg",
      "bootFromVolume": true,
      "rootVolumeGB": 40
    },
    "created_at": "2025-08-27T19:27:52.080758Z"
  }
]
```

#### 2.2 과목 등록
```http
POST /courses
```

**요청 본문:**
```json
{
  "course_id": "NEW-2025-1",
  "title": "새로운 구조 수업",
  "department": "컴퓨터공학과",
  "semester": "2025-1",
  "start_at": "2025-03-01",
  "end_at": "2025-06-30",
  "quota_profile": {
    "instances": 2,
    "cores": 4,
    "ramMB": 8192,
    "volumes": 2,
    "gigabytes": 100,
    "ports": 2,
    "floatingIPs": 1,
    "snapshots": 0
  },
  "defaults": {
    "bootFromVolume": true
  }
}
```

### 3. 수강 관리 (Enrollment Management)

#### 3.1 수강 등록
```http
POST /students/{student_id}/enroll
```

**요청 본문:**
```json
{
  "course_id": "AI-2025-1",
  "status": "active"
}
```

**응답 예시:**
```json
{
  "student_id": "32210003",
  "course_id": "AI-2025-1",
  "status": "active",
  "start_at": "2025-03-01T00:00:00Z",
  "end_at": "2025-06-30T00:00:00Z"
}
```

#### 3.2 수강 철회
```http
DELETE /students/{student_id}/enroll/{course_id}
```

**응답 예시:**
```json
{
  "message": "unenrolled successfully"
}
```

#### 3.3 수강 목록 조회
```http
GET /students/{student_id}/enrollments
```

**응답 예시:**
```json
{
  "student_id": "32210006",
  "name": "이지수",
  "email": "jisoo@dankook.ac.kr",
  "department": "소프트웨어학과",
  "keystone_user_id": "7d2339456f97447695ae8a5cbbbb2f61",
  "keystone_project_id": "a8fee14aff1d4072945d63b483394d9b",
  "created_at": "2025-08-28T03:14:48.472901Z"
}
```

### 4. 쿼타 관리 (Quota Management)

#### 4.1 현재 쿼타 조회
```http
GET /quota/current?projectId={project_id}
```

**응답 예시:**
```json
{
  "projectId": "a8fee14aff1d4072945d63b483394d9b",
  "nova": {
    "cores": {
      "limit": 18,
      "in_use": 0
    },
    "ramMB": {
      "limit": 36864,
      "in_use": 0
    },
    "instances": {
      "limit": 15,
      "in_use": 0
    }
  },
  "cinder": {
    "gigabytes": {
      "limit": 330,
      "in_use": 0
    },
    "volumes": {
      "limit": 15,
      "in_use": 0
    },
    "snapshots": {
      "limit": 10,
      "in_use": 0
    }
  },
  "neutron": {
    "port": {
      "limit": 15,
      "in_use": 0
    },
    "floatingIP": {
      "limit": 8,
      "in_use": 0
    }
  }
}
```

#### 4.2 프로파일 기반 쿼타 적용
```http
POST /quota/applyProfile
```

**요청 본문:**
```json
{
  "projectId": "a8fee14aff1d4072945d63b483394d9b",
  "profile": "basic",
  "dryRun": false,
  "includeDiff": true
}
```

**응답 예시:**
```json
{
  "projectId": "a8fee14aff1d4072945d63b483394d9b",
  "profile": "basic",
  "plan": {
    "instances": 10,
    "cores": 8,
    "ramMB": 16384,
    "volumes": 10,
    "gigabytes": 100,
    "ports": 10,
    "floatingIPs": 5,
    "snapshots": 10
  },
  "applied": true,
  "dryRun": false,
  "current": {
    "nova": { ... },
    "cinder": { ... }
  }
}
```

### 5. 리콘실 (Reconciliation)

#### 5.1 대량 리콘실
```http
POST /reconciliation/bulk
```

**응답 예시:**
```json
{
  "total_students": 16,
  "success_count": 7,
  "failed_count": 6,
  "pending_count": 3,
  "summary": "Reconciliation completed: 7 success, 6 failed, 3 pending",
  "student_results": [
    {
      "student_id": "32210006",
      "student_name": "이지수",
      "baseline_quota": { ... },
      "active_courses": [ ... ],
      "effective_quota": { ... },
      "applied_quota": { ... },
      "status": "success"
    }
  ]
}
```

#### 5.2 리콘실 상태 확인
```http
GET /reconciliation/status
```

**응답 예시:**
```json
{
  "status": "ready",
  "message": "Reconciliation service is ready",
  "timestamp": "2025-08-28T03:27:47.930195Z"
}
```

### 6. OpenStack 프로젝트 관리

#### 6.1 프로젝트 목록 조회
```http
GET /openstack/projects
```

**응답 예시:**
```json
{
  "projects": [
    {
      "id": "98077908079746cab05e896ebf54258f",
      "name": "student-32210003-project",
      "description": "Project for student 이지수 (32210003)",
      "enabled": true,
      "domain_id": "default"
    }
  ],
  "count": 1
}
```

#### 6.2 특정 학생 프로젝트 조회
```http
GET /openstack/projects/{student_id}
```

**응답 예시:**
```json
{
  "is_domain": false,
  "description": "Project for student 이지수 (32210003)",
  "domain_id": "default",
  "enabled": true,
  "id": "98077908079746cab05e896ebf54258f",
  "name": "student-32210003-project",
  "parent_id": "default"
}
```

### 7. 시스템 상태

#### 7.1 헬스체크
```http
GET /healthz
```

**응답 예시:**
```json
{
  "ok": true
}
```

## 🔐 인증 및 권한

### OpenStack Keystone 연동
- **인증 방식**: Token-based authentication
- **권한**: Admin 권한으로 프로젝트/사용자 생성
- **도메인**: Default 도메인 사용

### API 보안
- **Rate Limiting**: 현재 미구현
- **API Key**: 현재 미구현
- **CORS**: 현재 미구현

## 📊 에러 코드

| HTTP 상태 코드 | 에러 코드 | 설명 |
|---------------|-----------|------|
| 400 | BAD_REQUEST | 잘못된 요청 형식 |
| 404 | NOT_FOUND | 리소스를 찾을 수 없음 |
| 500 | INTERNAL_ERROR | 서버 내부 오류 |
| 502 | BAD_GATEWAY | OpenStack 연동 오류 |

## 🧪 테스트 예시

### cURL을 사용한 테스트

```bash
# 1. 서버 상태 확인
curl -X GET http://localhost:8080/healthz

# 2. 학생 등록
curl -X POST http://localhost:8080/students \
  -H "Content-Type: application/json" \
  -d '{"student_id": "TEST001", "name": "테스트학생", "email": "test@university.ac.kr", "department": "테스트학과"}'

# 3. 과목 등록
curl -X POST http://localhost:8080/courses \
  -H "Content-Type: application/json" \
  -d '{"course_id": "TEST-2025-1", "title": "테스트 과목", "department": "테스트학과", "semester": "2025-1", "start_at": "2025-03-01", "end_at": "2025-06-30", "quota_profile": {"instances": 1, "cores": 2, "ramMB": 4096, "volumes": 1, "gigabytes": 30, "ports": 1, "floatingIPs": 0, "snapshots": 0}}'

# 4. 수강 등록
curl -X POST "http://localhost:8080/students/TEST001/enroll" \
  -H "Content-Type: application/json" \
  -d '{"course_id": "TEST-2025-1", "status": "active"}'

# 5. 쿼타 확인
curl -X GET "http://localhost:8080/quota/current?projectId=TEST001"

# 6. 대량 리콘실
curl -X POST http://localhost:8080/reconciliation/bulk
```

## 📝 참고사항

1. **날짜 형식**: ISO 8601 형식 사용 (`YYYY-MM-DDTHH:MM:SSZ`)
2. **시간대**: UTC 기준
3. **ID 형식**: 문자열 (학생 ID, 과목 ID 등)
4. **쿼타 단위**: 
   - RAM: MB (메가바이트)
   - 디스크: GB (기가바이트)
   - 코어: 개수
   - 인스턴스/볼륨/포트: 개수

---

**문서 버전**: 1.0  
**최종 업데이트**: 2025년 8월 28일  
**작성자**: 이지수 (Lee Ji-su)
