# url-shortener

[English](./README.md) | [한국어](./README.ko.md)

Go 표준 라이브러리와 SQLite로만 만든 미니멀 URL 단축기 — 웹 프레임워크도, ORM도 안 씀.

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![SQLite](https://img.shields.io/badge/DB-SQLite-003B57?logo=sqlite&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

## 기능

- 긴 URL을 7자리 코드로 단축
- 단축 코드로 접속하면 원본 URL로 리다이렉트
- 같은 URL 재요청 시 새로 안 만들고 기존 단축 코드 재사용
- URL 형식 유효성 검증

## 기술 스택

- **Go** — 표준 라이브러리 `net/http`, 프레임워크 없음
- **SQLite** — [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) 사용, 순수 Go 드라이버(CGO 불필요)
- **Docker / Docker Compose** — 격리된 개발 환경

## 시작하기

### 필요한 것

- Docker Desktop

### 실행

```bash
git clone https://github.com/dev1f965x/url-shortener.git
cd url-shortener
docker compose up --build
```

서버는 `http://localhost:8080`에서 시작됩니다.

## API

### URL 단축

```
POST /shorten
Content-Type: application/json

{"url": "https://example.com/very/long/path"}
```

새 URL이면 `201 Created`, 이미 단축한 적 있는 URL이면 `200 OK`:

```json
{
  "short_code": "q1km5u2",
  "short_url": "http://localhost:8080/q1km5u2",
  "original_url": "https://example.com/very/long/path"
}
```

URL이 없거나 형식이 잘못됐으면 `400 Bad Request`.

### 리다이렉트

```
GET /{short_code}
```

원본 URL로 `302 Found` + `Location` 헤더 응답, 존재하지 않는 코드면 `404 Not Found`.

## 데이터 모델

```sql
CREATE TABLE urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_url TEXT NOT NULL UNIQUE,
    short_code TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 로드맵

- [ ] 커스텀 슬러그
- [ ] 클릭 수 카운트
- [ ] 링크 만료
- [ ] 링크 삭제/비활성화
- [ ] QR 코드 생성
- [ ] 클릭 통계 대시보드
- [ ] 계정 시스템
- [ ] API 키 발급

## 라이선스

[MIT](./LICENSE)
