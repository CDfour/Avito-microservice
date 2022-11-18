FROM golang

WORKDIR C:\go-work\Avito

COPY . .

EXPOSE 9000

CMD ["go","run","cmd/main.go"]