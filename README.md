# KubeDeploy Backend

Минимальный backend, предоставляющий только /health

## Проверка

```bash
make test
make build
```

## Контейнер

```bash
make image
docker run --rm -p 18080:8080 kubedeploy-backend:v0.1.0
curl --fail http://127.0.0.1:18080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```
