# check=skip=InvalidDefaultArgInFrom
ARG GO_VERSION
FROM golang:${GO_VERSION}

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

ENTRYPOINT ["go", "build", "-trimpath", "-o", "/out/nemu", "./cmd/nemu"]
