# syntax=docker/dockerfile:1
FROM base-image:latest
WORKDIR /workspace
COPY source.txt target.txt
RUN echo "ready"
CMD ["program", "argument"]
