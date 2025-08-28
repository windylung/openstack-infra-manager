# (2) Keystone 인증 검증

## 1. 환경 확인

### 1-1. VM 연결 확인

```bash
#VM 내에서 
ip adress

#로컬에서 
ping [ip 주소]
```

![UTM 내 VM ]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/8d23ff59-8c40-41bb-88aa-88a0d2ba087c.png)

UTM 내 VM 

![로컬]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/KakaoTalk_Photo_2025-08-19-00-54-56.png)

로컬

### 1-2. Keystone 접근 확인

```bash
curl http://${vm ip 주소}/identity/v3
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image.png)

- [`http://192.168.64.4:5000/v3`](http://192.168.64.4:5000/v3로는) 로는 접근이 안되는 것을 확인할 수 있다.
- Keystone이 standalone으로 띄워지면 <host>:5000/v3 이나,
DevStack은 Keystone을 /identity로 매핑함.

## 2. 배경

### 2-1. DevStack의 Apache 사용

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%201.png)

**(1) mod_wsgi**

- Apache(httpd) 안에서 동작하는 C 모듈
- OpenStack의 API 서비스들은 전부 Python WSGI 앱 → Python WSGI 애플리케이션 실행하는 역할

**(2) 동작 순서** 

1. Apache가 요청 수신
2. mod_wsgi가 “어떤 WSGI 앱에 매핑돼 있는지” 확인
3. Python 인터프리터 안에서 해당 WSGI callable 실행
4. 결과를 다시 Apache에 전달
5. 클라이언트에 HTTP 응답

## 3. 개요

## 4. 진행 과정

### 4-1. OpenStack 관리자 계정 환경 변수 확인

```bash
cd ~/devstack/
source ./openrc admin admin

echo "$OS_USERNAME"
echo "$OS_PROJECT_NAME"
echo "$OS_USER_DOMAIN_ID"
echo "$OS_PROJECT_DOMAIN_ID"
echo "$OS_PASSWORD"
```

### 4-2. 언스코프 토큰 발급

**(0) 언스코프, 스코프 토큰**

- **언스코프 토큰 (Unscoped)**
    - “사용자 본인 확인”만 증명.
    - `/auth/projects` 같은 **프로젝트 목록 조회**에 사용.
    - **프로젝트/도메인 스코프 정보가 없고**, 보통 **service catalog(엔드포인트 목록)도 없음**.
    - Nova/Cinder 같은 **프로젝트 자원 API 호출 불가**.
- **스코프 토큰 (Scoped)**
    - “사용자 + 특정 프로젝트(또는 도메인/시스템)에서의 권한” 포함.
    - 토큰 본문에 `project`와 `roles`가 들어가고, **service catalog**가 포함됨.
    - **실제 API 호출(Nova/Cinder 등)** 은 이 토큰으로만 가능.

**(1) 엔드포인트 응답 확인**

```bash
curl -s -o /dev/null -w '%{http_code}\n' http://192.168.64.4/identity/v3
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%202.png)

**(2) 언스코프 토큰 요청**

- user 정보는 4-1. OpenStack 관리자 계정 환경 변수 확인 에서 확인한 값을 입력해야 한다.

```bash
curl -si -H 'Content-Type: application/json' \
-d '{
  "auth": {
    "identity": {
      "methods": ["password"],
      "password": {
        "user": {
          "name": "admin",
          "domain": { "id": "default" },
          "password": "secret"
        }
      }
    }
  }
}' http://192.168.64.4/identity/v3/auth/tokens
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%203.png)

응답 결과 내 X-Subject-Token가 언스코프 토큰

**(3) 프로젝트 목록 조회**

```bash
curl -s -H "X-Auth-Token: $UNSCOPED_TOKEN" \
  http://192.168.64.4/identity/v3/auth/projects | jq .
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%204.png)

두 프로젝트(admin, demo)에 대해 토큰을 스코프할 자격이 있으나, **어떤 권한(roles)** 인지는 이 응답만으로는 알 수 없음

**(4) 프로젝트 내 권한 조회**

```bash
openstack role assignment list --user admin --project demo --names
# (3)에서 확인한 project의 name 입력
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%205.png)

### 4-3. 스코프 토큰 발급

**(1) project id로 토큰 발급** 

```bash
curl -si -H 'Content-Type: application/json' \
-d '{
  "auth": {
    "identity": {
      "methods": ["password"],
      "password": {
        "user": {
          "name": "admin",
          "domain": { "id": "default" },
          "password": "secret"
        }
      }
    },
    "scope": {
      "project": { "id": "144ba567a5db4334bce43a8c3b54f710" }
    }
  }
}' http://192.168.64.4/identity/v3/auth/tokens

#4-2-(3)에서 확인한 프로젝트의 id 사용 
```

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%206.png)

응답 결과 내 X-Subject-Token이 스코프 토큰 → 전반부는 언스코프 토큰과 동일할 수 있으나, 다른 토큰 값이다

**(2) 스코프 토큰 환경변수 설정** 

```bash
SCOPED_TOKEN=$(curl -si -H 'Content-Type: application/json' \
-d '{
  "auth": {
    "identity": { "methods": ["password"], "password": { "user": {
      "name": "admin", "domain": { "id": "default" }, "password": "secret"
    }}},
    "scope": { "project": { "id": "144ba567a5db4334bce43a8c3b54f710" } }
  }
}' http://192.168.64.4/identity/v3/auth/tokens \
| awk -F': ' '/^X-Subject-Token:/ {print $2}' | tr -d '\r')
echo "$SCOPED_TOKEN"
```

**(3) 스코프 토큰 확인**

```bash
curl -s   -H "X-Auth-Token: $SCOPED_TOKEN"   -H "X-Subject-Token: $SCOPED_TOKEN"   [http://192.168.64.4/identity/v3/auth/tokens](http://192.168.64.4/identity/v3/auth/tokens) | jq .
```

- 스코프 토큰의 경우, 특정 프로젝트에 대한 권한 포함(본문에 `project`, `roles`, `catalog`가 나타남)

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/image%207.png)

![image.png]((2)%20Keystone%20%EC%9D%B8%EC%A6%9D%20%EA%B2%80%EC%A6%9D%20253e19595aec80819c1dfcdf9b0b70e4/d28ffa3b-b58e-4f8b-9d8e-39df77530c9a.png)