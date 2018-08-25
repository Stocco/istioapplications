FROM golang:latest
ADD app /app
CMD ["/app"]
EXPOSE 8080
EXPOSE 8333