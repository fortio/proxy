FROM scratch
COPY proxy /usr/bin/proxy
ENTRYPOINT ["/usr/bin/proxy"]
EXPOSE 443
EXPOSE 80
# configmap (dynamic flags)
VOLUME /etc/fortio-proxy-config
# certs
VOLUME /etc/fortio-proxy-certs
# logs etc
WORKDIR /var/log/fortio
# start the proxy with default; the routes and cert by default
CMD ["-config", "/etc/fortio-proxy-config", "-certs-directory", "/etc/fortio-proxy-certs"]
