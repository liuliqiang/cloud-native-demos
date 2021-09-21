FROM golang:1.16

FROM busybox:glibc
COPY --from=0 /lib/x86_64-linux-gnu/libdl.so.2 /lib/libdl.so.2
COPY .build/server /home/liqiang/
COPY .build/client /home/liqiang/
RUN chmod +x /home/liqiang/*
WORKDIR /home/liqiang

# PORT
EXPOSE 2000-10000

CMD ["/home/liqiang/server"]