FROM openjdk:11-jre-slim
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Xms350m", "-Xmx350m", "-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]