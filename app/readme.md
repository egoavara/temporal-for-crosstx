# 실행법

kubectl port-forward를 이용해 PG, Temporal-Frontend를 노출시킵니다. (PG 5432:5432, Tempora-frontend 7233:7233)

```
go run main.go
```

```
curl http://localhost:9000/data/hello
curl http://localhost:9000/data/hello/current

curl http://localhost:9000/data/hello/update?patch=world
curl http://localhost:9000/data/hello/current

curl http://localhost:9000/data/hello/update?patch=loremipsum
curl http://localhost:9000/data/hello/current

curl http://localhost:9000/data/hello/commit
```