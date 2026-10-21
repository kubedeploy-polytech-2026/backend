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

## CI и релиз образа

Push в ветку и PR запускают unit-тесты, сборку контейнера, runtime smoke и
Trivy на нативных amd64 и arm64 runners. Исправимые HIGH/CRITICAL уязвимости
останавливают проверку; отчёт доступен в Actions artifacts.

После мерджа этого workflow в `main` создайте тег на текущем коммите `main`:

```bash
git switch main
git pull --ff-only
git tag v0.1.0
git push origin v0.1.0
```

Тег должен иметь вид `vMAJOR.MINOR.PATCH`. Workflow отклоняет тег на коммите,
который не совпадает с текущим `origin/main`. Тег предназначен для одного
релиза: опубликованный тег не следует переносить или переиспользовать.

Обе архитектуры сначала проходят сборку, runtime smoke и Trivy. Только после
успеха обеих проверок publish job получает право записи в GHCR и публикует
те же проверенные образы, без повторной сборки. Общий multi-platform образ:

```text
ghcr.io/kubedeploy-polytech-2026/backend:v0.1.0
ghcr.io/kubedeploy-polytech-2026/backend:sha-<полный SHA коммита>
```

Технические теги `v0.1.0-amd64` и `v0.1.0-arm64` содержат соответствующие
platform images. `latest` не публикуется. GitHub Release содержит
`release.json` с revision и digest общего образа, CycloneDX SBOM и Trivy-отчёт
для каждой архитектуры. Provenance подписывается для общего digest, SBOM —
для digest соответствующей архитектуры. Проверка provenance:

```bash
gh attestation verify oci://ghcr.io/kubedeploy-polytech-2026/backend@sha256:<digest> \
  --repo kubedeploy-polytech-2026/backend \
  --signer-workflow kubedeploy-polytech-2026/backend/.github/workflows/release.yml
```

Проверить контейнер локально:

```bash
make image
IMAGE=kubedeploy-backend:v0.1.0 bash scripts/test-image.sh
```
