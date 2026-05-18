FROM base-image
HEALTHCHECK --interval=30s --timeout=5s \
  CMD test -f /tmp/healthy || exit 1
HEALTHCHECK NONE
EXPOSE 8080/tcp
