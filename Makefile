IMG ?= ghostbaby/cfs-broker:v0.0.2
TARGET = cfs-broker

all: $(TARGET)

$(TARGET):
	go build -ldflags "-s -w" -o bin/$@

init:
	swag init

run:
	go run main.go -config config.json

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cfs-broker
	upx cfs-broker
	docker  build -t $(IMG) .

push:
	docker push $(IMG)

ci: build push

cd:
	kubectl config use-context aliqa
	kubectl apply -f ./contrib

del:
	kubectl config use-context aliqa
	kubectl delete -f ./contrib

roll:
	kubectl config use-context aliqa
	kubectl -n cfs patch ds cfs-broker --patch "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"

deploy: ci cd roll

rollout:ci roll

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

test: fmt vet
	go test ./... -coverprofile cover.out

TARGET = phantom-runtime
all: $(TARGET)

$(TARGET):
	go build -ldflags "-s -w" -o bin/$@
