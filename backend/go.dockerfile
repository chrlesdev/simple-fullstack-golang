FROM golang:1.27.3

WORKDIR /app

COPY . .

# Download and install the Depedencies

RUN go get -d -v ./...

# build the go app
RUN go build -o api .

EXPOSE 8000 

CMD ["./api"]