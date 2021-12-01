FROM openjdk:11-jre-slim
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Xms130m", "-Xmx130m", "-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]