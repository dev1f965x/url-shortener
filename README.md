# url-shortener

[English](./README.md) | [한국어](./README.ko.md)

URL shortener built on Go's standard library and SQLite.

## Run

```bash
docker compose up --build
```

The API listens on `http://localhost:8080` and stores data in `./data/urls.db`.

| Variable | Default | Description |
|---|---|---|
| `ADDR` | `:8080` | Listen address |
| `DB_PATH` | `data/urls.db` | SQLite file |
| `BASE_URL` | Request host | Prefix for returned short links |

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

`201` created, `200` already shortened, `400` invalid URL.

### `GET /{short_code}`

`302` to the original URL, or `404`.

## License

[MIT](./LICENSE)
