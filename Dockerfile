FROM golang:1.19.4-bullseye

RUN apt-get update && apt-get install -y \
    liblpsolve55-dev \
    && rm -rf /var/lib/apt/lists/*

ENV CGO_CFLAGS="-I/usr/include/lpsolve" \
    CGO_LDFLAGS="-llpsolve55 -lm -ldl -lcolamd"

WORKDIR /app
ENV GOMEMLIMIT=1800MiB
ENV GOGC=off

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o profit

CMD ["./profit", "-iters", "0", "-exporter", "solution"]
