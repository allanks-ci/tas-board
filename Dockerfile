FROM scratch
EXPOSE 8080

WORKDIR /server
COPY static /server/static
COPY app /server/tas-board

ENTRYPOINT ["./tas-board"]
CMD [""]