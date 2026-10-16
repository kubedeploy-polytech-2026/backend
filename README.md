# KubeDeploy Backend

HTTP API демонстрационного приложения KubeDeploy.

Доступные методы:

- `GET /health` — проверка работоспособности процесса;
- `GET /api/version` — имя и версия сервиса для frontend.

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
curl --fail http://127.0.0.1:18080/api/version
```

Ожидаемый ответ:

```json
{"status":"ok"}
{"service":"kubedeploy-backend","version":"0.1.0"}
```
