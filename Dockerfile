FROM scratch

COPY bin/czat /czat

ENV ZDJECIA=/zdjecia
VOLUME /zdjecia
EXPOSE 8080

ENTRYPOINT ["/czat"]
