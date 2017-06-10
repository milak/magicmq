# Use a base image
FROM golang:1.8

RUN echo "Building"
ADD build.sh /
RUN chmod a+x build.sh
RUN ./build.sh
# Set the working directory to /app
#WORKDIR /

# Copy the current directory contents into the container at /app
#ADD docker/magicmq /
ADD configuration.json /
#ADD web /web

# Make port 80 and 8080 available to the world outside this container
#EXPOSE 80
#EXPOSE 1789

# Define environment variable
#ENV NAME World

# Run magicmq when the container launches
CMD ["/mmq", "-f", "configuration.json"]
#CMD ["echo","in the container"]