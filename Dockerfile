FROM golang:1.16.3-alpine3.13
WORKDIR /app
COPY go.mod go.sum /app/
#RUN echo "Truncating sum ..." && truncate -s0 go.sum
#RUN echo "Downloading mods ..." && go mod download
#RUN echo "Downloading new version of modules ..." && go get -v -u
#RUN echo "Removing unnecessary libraries ..." && go mod tidy
COPY . /app
RUN echo "Building module ..." && go build -o main
EXPOSE 11001
CMD ["./main"]