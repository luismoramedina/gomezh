FROM openjdk

ADD target/*.jar /app.jar
WORKDIR /
ENV SERVER_PORT=8081
EXPOSE 8081
ENTRYPOINT ["java", "-jar", "app.jar"]