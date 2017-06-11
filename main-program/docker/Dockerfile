# Use a base image
FROM golang:alpin

ENV GOBIN /go/bin
WORKDIR /go/src
RUN go-wrapper download github.com/google/uuid
RUN go-wrapper download github.com/milak/tools
RUN go-wrapper download github.com/milak/mmq
RUN go-wrapper install github.com/milak/mmq/mmq/mmq.go

# Copy the current directory contents into the container at /app
ADD docker/configuration.json /
ADD web /web

# Make port 80 and 8080 available to the world outside this container
EXPOSE 80
EXPOSE 1789

# Define environment variable
#ENV NAME World
WORKDIR /
# Run magicmq when the container launches
CMD ["/go/bin/mmq", "-f", "configuration.json"]
#CMD ["echo","in the container"]
