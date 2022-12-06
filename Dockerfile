# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY datafiles ./datafiles
COPY device ./device
COPY genheap ./genheap
COPY stats ./stats

RUN go build -o /ts-query-workers

ENTRYPOINT [ "/ts-query-workers" ]