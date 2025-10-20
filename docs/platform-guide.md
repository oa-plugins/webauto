# webauto 플랫폼별 설치 가이드

**버전**: 1.0.0
**최종 업데이트**: 2025-10-20

이 문서는 webauto 플러그인의 플랫폼별(Windows, macOS, Linux) 설치 및 설정 가이드입니다.

---

## 목차

1. [지원 플랫폼](#지원-플랫폼)
2. [Windows 10/11 설치](#windows-1011-설치)
3. [macOS 설치](#macos-설치)
4. [Linux (Ubuntu) 설치](#linux-ubuntu-설치)
5. [공통 설정](#공통-설정)
6. [트러블슈팅](#트러블슈팅)

---

## 지원 플랫폼

| 플랫폼 | 아키텍처 | 브라우저 | 상태 |
|--------|---------|---------|------|
| **Windows 10/11** | amd64 (x86_64) | Chromium, Firefox | ✅ 완전 지원 |
| **macOS 11+** | amd64 (Intel) | Chromium, Firefox, WebKit | ✅ 완전 지원 |
| **macOS 11+** | arm64 (Apple Silicon) | Chromium, Firefox, WebKit | ✅ 완전 지원 |
| **Ubuntu 20.04+** | amd64 (x86_64) | Chromium, Firefox | ✅ 완전 지원 |
| **Ubuntu 20.04+** | arm64 | Chromium, Firefox | ⚠️  제한적 지원 |

**참고**: WebKit은 Linux에서 제한적으로 지원됩니다.

---

## Windows 10/11 설치

### 1. 사전 요구사항

**필수 소프트웨어**:
- **Go 1.22+**: [공식 다운로드](https://go.dev/dl/)
- **Git for Windows**: [공식 다운로드](https://git-scm.com/download/win)

**Node.js**: webauto는 Node.js를 자동으로 다운로드하고 설치하므로 **별도 설치가 필요 없습니다**.

### 2. webauto 설치

#### 방법 A: OA CLI를 통한 설치 (권장)

```powershell
# OA CLI 설치 (아직 설치하지 않은 경우)
# [OA 저장소](https://github.com/oa-plugins/oa) 참조

# webauto 플러그인 설치
oa plugin install webauto
```

#### 방법 B: 직접 빌드

```powershell
# 1. 저장소 클론
git clone https://github.com/oa-plugins/webauto.git
cd webauto

# 2. 빌드
go build -o webauto.exe cmd/webauto/main.go

# 3. OA에 등록 (선택)
oa plugin install ./webauto.exe
```

### 3. 설치 확인

```powershell
# webauto 버전 확인
oa webauto --version

# 브라우저 시작 테스트 (Node.js 자동 부트스트랩)
oa webauto browser-launch --headless true

# 출력 예시:
# {"success":true,"data":{"session_id":"ses_xxx",...},"error":null,...}
```

### 4. Windows 특화 설정

#### 4.1 PowerShell 실행 정책

webauto는 PowerShell 스크립트를 사용하지 않지만, 일부 Node.js 패키지가 실행 정책을 요구할 수 있습니다:

```powershell
# 현재 사용자에 대해 실행 정책 변경 (관리자 권한 필요)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 4.2 방화벽 설정

Chromium/Firefox가 네트워크에 접근하려면 방화벽 허용이 필요할 수 있습니다. 첫 실행 시 Windows Defender 방화벽 프롬프트가 표시되면 **허용**을 선택하세요.

#### 4.3 Node.js 및 브라우저 캐시 위치

```
C:\Users\<사용자명>\AppData\Local\oa\webauto\
├── runtime/      # Node.js 런타임
├── node_modules/ # Playwright 패키지
└── browsers/     # 브라우저 바이너리 (Chromium, Firefox)
```

---

## macOS 설치

### 1. 사전 요구사항

**필수 소프트웨어**:
- **Go 1.22+**: [공식 다운로드](https://go.dev/dl/) 또는 Homebrew:
  ```bash
  brew install go
  ```
- **Git**: macOS에 기본 포함

**Node.js**: webauto는 Node.js를 자동으로 다운로드하고 설치하므로 **별도 설치가 필요 없습니다**.

### 2. webauto 설치

#### 방법 A: OA CLI를 통한 설치 (권장)

```bash
# OA CLI 설치 (아직 설치하지 않은 경우)
# [OA 저장소](https://github.com/oa-plugins/oa) 참조

# webauto 플러그인 설치
oa plugin install webauto
```

#### 방법 B: 직접 빌드

```bash
# 1. 저장소 클론
git clone https://github.com/oa-plugins/webauto.git
cd webauto

# 2. 빌드 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o webauto cmd/webauto/main.go

# 또는 Intel Mac
GOOS=darwin GOARCH=amd64 go build -o webauto cmd/webauto/main.go

# 3. 실행 권한 부여
chmod +x webauto

# 4. OA에 등록 (선택)
oa plugin install ./webauto
```

### 3. 설치 확인

```bash
# webauto 버전 확인
oa webauto --version

# 브라우저 시작 테스트 (Node.js 자동 부트스트랩)
oa webauto browser-launch --headless true

# 출력 예시:
# {"success":true,"data":{"session_id":"ses_xxx",...},"error":null,...}
```

### 4. macOS 특화 설정

#### 4.1 Apple Silicon (M1/M2/M3) 사용자

**Rosetta 2** (선택): Intel 바이너리 호환성을 위해 Rosetta 2를 설치할 수 있습니다:

```bash
softwareupdate --install-rosetta --agree-to-license
```

**참고**: webauto ARM64 바이너리는 Rosetta 2 없이 네이티브로 실행됩니다.

#### 4.2 Gatekeeper 권한

처음 실행 시 macOS Gatekeeper가 "인증되지 않은 개발자" 경고를 표시할 수 있습니다:

```bash
# 방법 1: 시스템 설정에서 수동 허용
# System Preferences → Security & Privacy → General → "Allow Anyway"

# 방법 2: 명령어로 격리 속성 제거
xattr -d com.apple.quarantine ./webauto
```

#### 4.3 Node.js 및 브라우저 캐시 위치

```
/Users/<사용자명>/.cache/oa/webauto/
├── runtime/      # Node.js 런타임
├── node_modules/ # Playwright 패키지
└── browsers/     # 브라우저 바이너리 (Chromium, Firefox, WebKit)
```

#### 4.4 WebKit 지원

macOS는 WebKit (Safari 엔진)을 완전히 지원합니다. WebKit 테스트를 위해 추가 설치가 필요 없습니다.

---

## Linux (Ubuntu) 설치

### 1. 사전 요구사항

**필수 소프트웨어**:
- **Go 1.22+**:
  ```bash
  # 방법 1: 공식 바이너리 다운로드
  wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  source ~/.bashrc

  # 방법 2: apt (Ubuntu 22.04+)
  sudo apt update
  sudo apt install golang-1.22
  ```

- **Git**:
  ```bash
  sudo apt update
  sudo apt install git
  ```

**Node.js**: webauto는 Node.js를 자동으로 다운로드하고 설치하므로 **별도 설치가 필요 없습니다**.

### 2. 시스템 종속성 설치

Playwright 브라우저가 Linux에서 실행되려면 추가 라이브러리가 필요합니다:

```bash
# Ubuntu/Debian 계열
sudo apt update
sudo apt install -y \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxrandr2 \
    libgbm1 \
    libasound2
```

**헤드리스 서버 (X11 없음)**:

```bash
# 가상 디스플레이 서버 (선택)
sudo apt install xvfb

# webauto를 Xvfb로 실행
xvfb-run oa webauto browser-launch --headless true
```

### 3. webauto 설치

#### 방법 A: OA CLI를 통한 설치 (권장)

```bash
# OA CLI 설치 (아직 설치하지 않은 경우)
# [OA 저장소](https://github.com/oa-plugins/oa) 참조

# webauto 플러그인 설치
oa plugin install webauto
```

#### 방법 B: 직접 빌드

```bash
# 1. 저장소 클론
git clone https://github.com/oa-plugins/webauto.git
cd webauto

# 2. 빌드
GOOS=linux GOARCH=amd64 go build -o webauto cmd/webauto/main.go

# 3. 실행 권한 부여
chmod +x webauto

# 4. OA에 등록 (선택)
oa plugin install ./webauto
```

### 4. 설치 확인

```bash
# webauto 버전 확인
oa webauto --version

# 브라우저 시작 테스트 (Node.js 자동 부트스트랩)
oa webauto browser-launch --headless true

# 출력 예시:
# {"success":true,"data":{"session_id":"ses_xxx",...},"error":null,...}
```

### 5. Linux 특화 설정

#### 5.1 권한 문제

```bash
# webauto 실행 파일에 실행 권한 부여
chmod +x webauto

# 캐시 디렉토리 권한 확인
ls -ld ~/.cache/oa/webauto
```

#### 5.2 Node.js 및 브라우저 캐시 위치

```
/home/<사용자명>/.cache/oa/webauto/
├── runtime/      # Node.js 런타임
├── node_modules/ # Playwright 패키지
└── browsers/     # 브라우저 바이너리 (Chromium, Firefox)
```

#### 5.3 WebKit 지원 (제한적)

Linux에서 WebKit 지원은 제한적입니다. Chromium 또는 Firefox 사용을 권장합니다.

---

## 공통 설정

### 1. 환경 변수 (선택)

webauto는 다음 환경 변수를 지원합니다:

```bash
# Anti-Bot 전략 활성화 (기본값: true)
export ENABLE_STEALTH=true
export ENABLE_FINGERPRINT=true
export ENABLE_BEHAVIOR_RANDOM=true

# 행동 패턴 랜덤화 설정
export TYPING_DELAY_MS=30          # 타이핑 지연 (밀리초)
export MOUSE_MOVE_JITTER_PX=10     # 마우스 이동 지터 (픽셀)

# 브라우저 타임아웃 설정
export BROWSER_LAUNCH_TIMEOUT_MS=30000
export PAGE_LOAD_TIMEOUT_MS=30000
export ELEMENT_TIMEOUT_MS=5000
```

**설정 방법**:

- **Windows (PowerShell)**:
  ```powershell
  $env:ENABLE_STEALTH = "true"
  ```

- **macOS/Linux (Bash)**:
  ```bash
  echo 'export ENABLE_STEALTH=true' >> ~/.bashrc
  source ~/.bashrc
  ```

### 2. 브라우저 선택

```bash
# Chromium (기본값)
oa webauto browser-launch --browser-type chromium

# Firefox
oa webauto browser-launch --browser-type firefox

# WebKit (macOS만 완전 지원)
oa webauto browser-launch --browser-type webkit
```

### 3. 헤드리스 모드 vs. 헤드풀 모드

```bash
# 헤드리스 모드 (백그라운드 실행, 기본값)
oa webauto browser-launch --headless true

# 헤드풀 모드 (브라우저 UI 표시)
oa webauto browser-launch --headless false
```

---

## 트러블슈팅

### 문제 1: "NODE_NOT_FOUND" 에러

**증상**:
```json
{
  "success": false,
  "error": {
    "code": "NODE_NOT_FOUND",
    "message": "Node.js runtime not found or not bootstrapped"
  }
}
```

**해결 방법**:

1. **Node.js 자동 부트스트랩 재시도**:
   ```bash
   # 캐시 삭제 후 재시도
   rm -rf ~/.cache/oa/webauto  # macOS/Linux
   # 또는
   Remove-Item -Recurse -Force "$env:LOCALAPPDATA\oa\webauto"  # Windows

   # 다시 실행
   oa webauto browser-launch --headless true
   ```

2. **수동 Node.js 설치** (대안):
   - Node.js 18+를 시스템에 설치하고 PATH에 추가
   - [공식 다운로드](https://nodejs.org/)

### 문제 2: "PLAYWRIGHT_NOT_INSTALLED" 에러

**증상**:
```json
{
  "success": false,
  "error": {
    "code": "PLAYWRIGHT_NOT_INSTALLED",
    "message": "Playwright package not installed"
  }
}
```

**해결 방법**:

```bash
# 캐시 디렉토리로 이동
cd ~/.cache/oa/webauto  # macOS/Linux
# 또는
cd "$env:LOCALAPPDATA\oa\webauto"  # Windows

# Playwright 수동 설치
npm install playwright
npx playwright install chromium firefox webkit
```

### 문제 3: Linux에서 브라우저 실행 실패

**증상**:
```
error while loading shared libraries: libnss3.so
```

**해결 방법**:

```bash
# 누락된 시스템 라이브러리 설치
sudo apt update
sudo apt install -y libnss3 libatk1.0-0 libatk-bridge2.0-0 libcups2 libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 libxrandr2 libgbm1 libasound2
```

### 문제 4: macOS Gatekeeper 경고

**증상**:
```
"webauto" cannot be opened because the developer cannot be verified.
```

**해결 방법**:

```bash
# 격리 속성 제거
xattr -d com.apple.quarantine ./webauto

# 또는 시스템 설정에서 수동 허용
# System Preferences → Security & Privacy → General → "Allow Anyway"
```

### 문제 5: Windows PowerShell 실행 정책

**증상**:
```
... cannot be loaded because running scripts is disabled on this system.
```

**해결 방법**:

```powershell
# 현재 사용자에 대해 실행 정책 변경 (관리자 권한 필요)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 문제 6: 브라우저 다운로드 타임아웃 (느린 네트워크)

**증상**:
```
Failed to download browser: timeout
```

**해결 방법**:

1. **타임아웃 증가**:
   ```bash
   export BROWSER_LAUNCH_TIMEOUT_MS=60000  # 60초
   oa webauto browser-launch
   ```

2. **수동 브라우저 다운로드**:
   ```bash
   cd ~/.cache/oa/webauto
   npx playwright install chromium
   ```

### 문제 7: Anti-Bot 탐지로 인한 차단

**증상**:
특정 웹사이트에서 "봇으로 의심되는 접근" 메시지

**해결 방법**:

1. **Anti-Bot 전략 활성화**:
   ```bash
   export ENABLE_STEALTH=true
   export ENABLE_FINGERPRINT=true
   export ENABLE_BEHAVIOR_RANDOM=true
   ```

2. **헤드풀 모드 사용**:
   ```bash
   oa webauto browser-launch --headless false
   ```

3. **타이핑 지연 증가**:
   ```bash
   export TYPING_DELAY_MS=100  # 100ms 지연
   ```

### 문제 8: "SESSION_NOT_FOUND" 에러

**증상**:
```json
{
  "success": false,
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session with ID 'ses_xxx' not found"
  }
}
```

**해결 방법**:

```bash
# 활성 세션 목록 확인
oa webauto session-list

# 새 세션 시작
oa webauto browser-launch --headless true
```

---

## 추가 리소스

- **GitHub 저장소**: [oa-plugins/webauto](https://github.com/oa-plugins/webauto)
- **Issue Tracker**: [GitHub Issues](https://github.com/oa-plugins/webauto/issues)
- **아키텍처 문서**: [ARCHITECTURE.md](../ARCHITECTURE.md)
- **구현 가이드**: [implementation-guide.md](implementation-guide.md)
- **Playwright 문서**: [playwright.dev](https://playwright.dev/)

---

## 문의 및 지원

- **버그 리포트**: [GitHub Issues](https://github.com/oa-plugins/webauto/issues)
- **기능 제안**: [GitHub Discussions](https://github.com/oa-plugins/webauto/discussions)

---

**마지막 업데이트**: 2025-10-20
**작성자**: webauto 개발팀
