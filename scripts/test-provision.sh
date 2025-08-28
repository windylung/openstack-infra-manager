#!/bin/bash

# OpenStack 프로비저닝 API 테스트 스크립트
# 사용법: ./scripts/test-provision.sh

set -e

# 기본 설정 (실제 환경에 맞게 수정 필요)
API_BASE="http://localhost:8080"
SERVER_NAME="vm-test-01"
IMAGE_ID="4a83c094-4e19-424c-9a80-83080fad7682"  # 실제 이미지 ID로 수정
FLAVOR_ID="1"  # m1.tiny의 ID (이름 아님)
NETWORK_ID="8c42b1b0-ecca-490c-9186-8be1c230fce2"  # 실제 네트워크 ID로 수정
KEY_NAME="my-key"  # 실제 키페어 이름으로 수정
EXTERNAL_NETWORK_ID="c6c2423f-a4be-40f9-89d5-17df2910fcf2"  # public 네트워크 ID

echo "🚀 OpenStack 프로비저닝 API 테스트 시작"
echo "=================================="

# 1. 헬스체크
echo "1. 서버 상태 확인..."
curl -s "$API_BASE/healthz" | jq '.'

# 2. 프로비저닝 요청
echo ""
echo "2. VM 프로비저닝 요청..."
echo "   서버명: $SERVER_NAME"
echo "   이미지: $IMAGE_ID"
echo "   플레이버: $FLAVOR_ID"
echo "   네트워크: $NETWORK_ID"
echo "   키페어: $KEY_NAME"

RESPONSE=$(curl -s "$API_BASE/provision/server" \
  -H 'Content-Type: application/json' \
  -d "{
    \"name\": \"$SERVER_NAME\",
    \"imageId\": \"$IMAGE_ID\",
    \"flavorId\": \"$FLAVOR_ID\",
    \"networkId\": \"$NETWORK_ID\",
    \"keyName\": \"$KEY_NAME\",
    \"securityGroups\": [\"default-ssh\"],
    \"assignFloatingIp\": true,
    \"externalNetworkId\": \"$EXTERNAL_NETWORK_ID\"
  }")

echo ""
echo "응답:"
echo "$RESPONSE" | jq '.'

# 3. 결과 확인
if echo "$RESPONSE" | jq -e '.serverId' > /dev/null 2>&1; then
    SERVER_ID=$(echo "$RESPONSE" | jq -r '.serverId')
    FLOATING_IP=$(echo "$RESPONSE" | jq -r '.floatingIp // empty')
    
    echo ""
    echo "✅ 프로비저닝 성공!"
    echo "   서버 ID: $SERVER_ID"
    if [ -n "$FLOATING_IP" ]; then
        echo "   플로팅 IP: $FLOATING_IP"
        echo "   SSH 접속: ssh ubuntu@$FLOATING_IP"
    fi
else
    echo ""
    echo "❌ 프로비저닝 실패"
    echo "   오류: $(echo "$RESPONSE" | jq -r '.error // "알 수 없는 오류"')"
    exit 1
fi

echo ""
echo "🎉 테스트 완료!"
