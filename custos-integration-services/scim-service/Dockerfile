FROM openjdk:11-jre-slim
COPY src/main/resources/custos_user_schema_extention.json /home/ubuntu/schemas/custos_user_schema_extention.json
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]
