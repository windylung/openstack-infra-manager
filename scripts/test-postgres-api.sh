#!/bin/bash

# PostgreSQL API 테스트 스크립트 (간소화 버전)
# 테스트할 API: 학생/수업/수강 기본 기능만

BASE_URL="http://localhost:8080"
SERVER_PID=""

# 서버 시작
start_server() {
    echo "🚀 서버 시작 중..."
    cd cmd/server
    go run . &
    SERVER_PID=$!
    cd ../..
    
    # 서버 시작 대기
    echo "⏳ 서버 시작 대기 중..."
    for i in {1..10}; do
        if curl -s "$BASE_URL/healthz" > /dev/null; then
            echo "✅ 서버 시작 완료!"
            return 0
        fi
        sleep 1
    done
    
    echo "❌ 서버 시작 실패"
    return 1
}

# 서버 종료
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        echo "🛑 서버 종료 중..."
        kill $SERVER_PID 2>/dev/null
        wait $SERVER_PID 2>/dev/null
    fi
}

# 에러 시 정리
trap cleanup EXIT

# 테스트 실행
run_tests() {
    echo "🧪 PostgreSQL API 테스트 시작"
    echo "=================================="
    
    # 1. 서버 상태 확인
    echo "1️⃣ 서버 상태 확인"
    response=$(curl -s "$BASE_URL/healthz")
    echo "   응답: $response"
    echo ""
    
    # 2. 수업 생성
    echo "2️⃣ 수업 생성"
    course_data='{
        "course_id": "CC-2025-1",
        "title": "클라우드 컴퓨팅",
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
            "floatingIPs": 1
        },
        "defaults": {
            "imageId": "ubuntu-20.04",
            "flavorIds": ["m1.medium", "m1.large"],
            "networkId": "private-network",
            "externalNetworkId": "public-network",
            "securityGroup": "default-sg",
            "bootFromVolume": true,
            "rootVolumeGB": 50
        }
    }'
    
    response=$(curl -s -X POST "$BASE_URL/courses" \
        -H "Content-Type: application/json" \
        -d "$course_data")
    echo "   응답: $response"
    echo ""
    
    # 3. 학생 등록
    echo "3️⃣ 학생 등록"
    student_data='{
        "student_id": "2024001",
        "name": "홍길동",
        "email": "hong@university.ac.kr",
        "department": "컴퓨터공학과"
    }'
    
    response=$(curl -s -X POST "$BASE_URL/students" \
        -H "Content-Type: application/json" \
        -d "$student_data")
    echo "   응답: $response"
    echo ""
    
    # 4. 수강 신청
    echo "4️⃣ 수강 신청"
    enrollment_data='{
        "course_id": "CC-2025-1",
        "status": "active",
        "start_at": "2025-03-01",
        "end_at": "2025-06-30"
    }'
    
    response=$(curl -s -X POST "$BASE_URL/students/2024001/enroll" \
        -H "Content-Type: application/json" \
        -d "$enrollment_data")
    echo "   응답: $response"
    echo ""
    
    # 5. 수강 현황 조회
    echo "5️⃣ 수강 현황 조회"
    response=$(curl -s "$BASE_URL/students/2024001/enrollments")
    echo "   응답: $response"
    echo ""
    
    # 6. 수업/학생 목록 조회
    echo "6️⃣ 수업 목록 조회"
    response=$(curl -s "$BASE_URL/courses")
    echo "   응답: $response"
    echo ""
    
    echo "7️⃣ 학생 목록 조회"
    response=$(curl -s "$BASE_URL/students")
    echo "   응답: $response"
    echo ""
    
    # 8. 수강 철회
    echo "8️⃣ 수강 철회"
    response=$(curl -s -X DELETE "$BASE_URL/students/2024001/enroll/CC-2025-1")
    echo "   응답: $response"
    echo ""
    
    # 9. 최종 상태 확인
    echo "9️⃣ 최종 상태 확인"
    response=$(curl -s "$BASE_URL/students/2024001/enrollments")
    echo "   응답: $response"
    echo ""
    
    echo "✅ 모든 테스트 완료!"
}

# 메인 실행
main() {
    if ! start_server; then
        exit 1
    fi
    
    run_tests
}

main "$@"
