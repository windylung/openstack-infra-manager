# Openstack Infra Management System

**학생별 과목 수강에 따른 OpenStack 자원 할당 및 쿼타 관리 시스템**

## 📋 프로젝트 개요

대학에서 학생들이 수강하는 과목에 따라 OpenStack 클라우드 자원을 자동으로 할당하고 관리하는 인프라 관리 시스템입니다.

## 🚀 주요 기능

- **학생 관리**: 학생 등록 및 OpenStack 프로젝트 자동 생성
- **과목 관리**: 과목별 쿼타 프로파일 정의 및 관리
- **수강 관리**: 수강 등록/철회 시 실시간 쿼타 조정
- **자원 할당**: OpenStack Nova, Cinder, Neutron 쿼타 자동 설정
- **리콘실**: 대량 및 개별 쿼타 조정 서비스

## 📚 문서

- [(1)DevStack 설치](./docs/(1)%20DevStack%20설치%20253e19595aec8099a57bc3000a9a0ffb.md)
- [(2)Keystone 인증 검증](./docs/(2)%20Keystone%20인증%20검증%20253e19595aec80819c1dfcdf9b0b70e4.md)
- [(3)Quota 설정-1](./docs/(3)%20Quota%20설정-1%20254e19595aec803c92d6da0a4cf0b53a.md)
- [(4)Quota 설정-2](./docs/(4)%20Quota%20설정-2%20254e19595aec80dea118ca065c41fb8d.md)
- [(5)자원관리: 준비](./docs/(5)%20자원%20관리%20준비%2025ce19595aec80b3b35edc7a28d15e5a.md)
- [(6)자원관리: 설계](./docs/(6)%20자원%20관리%20설계%2025ce19595aec80a98db4f2395a1e8d61.md)
- [(7)자원관리: 학기 시작 자원 초기 할당](./docs/(7)%20자원%20관리%20학기%20시작_자원%20초기%20할당%2025ce19595aec804291aee87c0f67a888.md)
- [API 명세서](./docs/api.md)


## 📁 프로젝트 구조

```
quota-api/
├── cmd/server/           # 메인 서버 진입점
├── internal/             # 내부 패키지
│   ├── api/             # API 타입 정의
│   ├── config/          # 설정 관리
│   ├── database/        # 데이터베이스 레이어
│   ├── http/            # HTTP 핸들러
│   ├── models/          # 데이터 모델
│   ├── openstack/       # OpenStack 연동
│   └── services/        # 비즈니스 로직 서비스
├── scripts/              # 테스트 및 유틸리티 스크립트
└── docs/                # 프로젝트 문서
```

## 🛠️ 기술 스택

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15
- **Cloud**: OpenStack (Keystone, Nova, Cinder, Neutron)
- **SDK**: Gophercloud v2
- **Container**: Docker



## 🚀 빠른 시작

### 1. 환경 설정

```bash
# 환경 변수 설정
export OS_AUTH_URL=http://your-openstack:5000/v3
export OS_USERNAME=your-username
export OS_PASSWORD=your-password
export OS_USER_DOMAIN_ID=default
export OS_PROJECT_NAME=your-project
export OS_REGION_NAME=your-region
```

### 2. 데이터베이스 실행

```bash
# PostgreSQL 컨테이너 실행
docker run -d --name postgres-quota \
  -e POSTGRES_DB=quota_db \
  -e POSTGRES_USER=quota_user \
  -e POSTGRES_PASSWORD=quota_password \
  -p 5432:5432 \
  postgres:15
```

### 3. 서버 실행

```bash
cd cmd/server
go run .
```

## 📊 현재 구현 상태

| 기능 | 상태 | 비고 |
|------|------|------|
| DevStack 설치 | ✅ 완료 | OpenStack 환경 구축 |
| Keystone 인증 검증 | ✅ 완료 | 사용자 인증 시스템 |
| Quota 설정-1 | ✅ 완료 | 기본 쿼타 관리 |
| Quota 설정-2 | ✅ 완료 | 고급 쿼타 기능 |
| 자원 관리 준비 | ✅ 완료 | 시스템 설계 |
| 자원 관리 설계 | ✅ 완료 | 아키텍처 설계 |
| 학기 시작 자원 초기 할당 | ✅ 완료 | 대량 리콘실 시스템 |

## 🚧 추후 개발 계획

- [ ] 스케줄러 기반 자동 수강 만료 처리
- [ ] 자원 회수 정책 (Nova, Neutron, Cinder)
- [ ] API 인증 시스템 (JWT 토큰)
- [ ] 관리자 웹 대시보드
- [ ] Docker 컨테이너화 및 CI/CD
- [ ] 모니터링 시스템 (Prometheus + Grafana)

## 🔗 API 엔드포인트

### 학생 관리
- `GET /students` - 학생 목록 조회
- `POST /students` - 학생 등록
- `GET /students/{id}` - 학생 상세 조회
### 과목 관리
- `GET /courses` - 과목 목록 조회
- `POST /courses` - 과목 등록
- `GET /courses/{id}` - 과목 상세 조회

### 수강 관리
- `POST /students/{id}/enroll` - 수강 등록
- `DELETE /students/{id}/enroll/{courseId}` - 수강 철회
- `GET /students/{id}/enrollments` - 수강 목록 조회

### 쿼타 관리
- `GET /quota/current?projectId={id}` - 현재 쿼타 조회
- `POST /quota/applyProfile` - 프로파일 기반 쿼타 적용
- `POST /reconciliation/bulk` - 대량 쿼타 리콘실



---

**개발자**: 이지수 @windylung
**최종 업데이트**: 2025년 8월 28일
