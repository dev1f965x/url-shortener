# url-shortener

[English](./README.md) | [한국어](./README.ko.md)

Go 표준 라이브러리와 SQLite로 만든 URL 단축기입니다.

## 실행

```bash
docker compose up --build
```

`http://localhost:8080`에서 실행되고 데이터는 `./data/urls.db`에 저장됩니다.

| 환경 변수 | 기본값 | 설명 |
|---|---|---|
| `ADDR` | `:8080` | 서버 주소 |
| `DB_PATH` | `data/urls.db` | SQLite 파일 경로 |
| `BASE_URL` | 요청 호스트 | 응답에 넣을 단축 링크 주소 |

## API

### `POST /shorten`

```json
{ "url": "https://example.com/long/path" }
```

```json
{
  "short_code": "q1km5u2",
  "short_url": "http://localhost:8080/q1km5u2",
  "original_url": "https://example.com/long/path"
}
```

새로 만들면 `201`, 이미 있는 URL이면 `200`, 형식이 잘못되면 `400`을 반환합니다.

### `GET /{short_code}`

원래 URL로 `302` 리다이렉트합니다. 없는 코드면 `404`입니다.

## 라이선스

[MIT](./LICENSE)
