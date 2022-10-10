FROM openjdk:11-jre-slim
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Xms250m", "-Xmx250m", "-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]