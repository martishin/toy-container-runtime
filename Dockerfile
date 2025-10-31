FROM alpine:3.20
RUN apk add --no-cache python3 bash busybox-suid
WORKDIR /opt/app
COPY app/hello.py /opt/app/hello.py
RUN chmod +x /opt/app/hello.py

CMD ["/usr/bin/python3", "/opt/app/hello.py"]
