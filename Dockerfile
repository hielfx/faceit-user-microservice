FROM go:1.19 AS BUILDER

WORKDIR /app

COPY . .

RUN go mod download && make


FROM alpine:3.15

WORKDIR /app

COPY --from=BUILDER /app/bin/users-microservice ./

EXPOSE 4040

CMD [ "./users-microservice" ]