FROM gcr.io/distroless/java:8
COPY /target/cert-manager-1.0-SNAPSHOT.jar /app/cert-manager.jar
WORKDIR /app
CMD ["cert-manager.jar"]
