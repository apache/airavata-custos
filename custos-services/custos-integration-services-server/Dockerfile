FROM openjdk:17.0.2-slim
COPY src/main/resources/custos_user_schema_extention.json /home/ubuntu/custos_user_schema_extention.json
VOLUME /tmp
ARG JAR_FILE
ADD ${JAR_FILE} app.jar
ENTRYPOINT ["java","-Xms130m", "-Xmx130m", "-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]