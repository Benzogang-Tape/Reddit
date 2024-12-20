FROM golang:1.23.3-alpine AS build_stage

LABEL authors="Benzogang-Tape"

ARG PORT
ARG APP_NAME
ARG BUILD_DATE

LABEL build_date=$BUILD_DATE

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /bin/$APP_NAME ./cmd/$APP_NAME

FROM alpine AS run_stage

RUN mkdir /app
WORKDIR /app
RUN mkdir /static
RUN mkdir /docs

COPY --from=build_stage /bin/$APP_NAME .

RUN chmod +x $APP_NAME

#EXPOSE $PORT
#
#ENTRYPOINT ["./$APP_NAME"]

EXPOSE $PORT

CMD [ "./$APP_NAME" ]
