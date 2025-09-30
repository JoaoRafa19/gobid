FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN go install github.com/jackc/tern@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

RUN apk add --no-cache make

COPY go.mod go.sum ./

COPY .env .

RUN go mod download

COPY . .

RUN make buildapp


FROM alpine:latest


WORKDIR /app

RUN apk add --no-cache postgresql bash

COPY --from=builder /app/bin/api .
COPY --from=builder /app/.env .
COPY --from=builder /go/bin/tern /usr/local/bin/tern
COPY --from=builder /app/internal/store/pgstore/migrations /app/internal/store/pgstore/migrations

RUN mkdir -p /app/scripts

COPY scripts/init.sh /app/scripts/init.sh
RUN chmod +x /app/scripts/init.sh

EXPOSE 3080

CMD [ "/app/scripts/init.sh" ]
