# terraform을 통한 temporal 구축 예시

```sh

terraform init

terraform apply



kubectl exec -i -t -n default {파드명} -c admin-tools -- sh -c "clear; (bash || ash || sh)"

temporal operator namespace create --namespace default
temporal operator search-attribute create --name Title --type Keyword

```