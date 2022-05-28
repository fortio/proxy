FROM scratch
COPY proxy /usr/bin/proxy
ENTRYPOINT ["/usr/bin/proxy"]
CMD ["-h"]
