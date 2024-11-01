FROM golang:1.22.0 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /spicy_dice ./cmd/main.go

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /spicy_dice /spicy_dice
COPY --from=build-stage /app/frontend /frontend

EXPOSE 8080

USER nonroot:nonroot

CMD ["/spicy_dice"]