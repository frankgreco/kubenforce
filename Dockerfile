FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
ADD kubenforce /
CMD ["/kubenforce"]
