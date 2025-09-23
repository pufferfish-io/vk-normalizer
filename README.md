# vk-normalizer

```
export $(cat .env | xargs) && go run ./cmd/vk-normalizer
```

```
go mod tidy
```

```
go build -v -x ./cmd/vk-normalizer && rm -f vk-normalizer
```

```
docker buildx build --no-cache --progress=plain .
```

```
set -a && source .env && set +a && go run ./cmd/vk-normalizer
```

```
git tag v0.0.1
git push origin v0.0.1
```

```
git tag -l
git tag -d v0.0.1
git push --delete origin v0.0.1
git ls-remote --tags origin | grep 'refs/tags/v0.0.1$'
```
