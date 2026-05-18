FROM base-image
RUN echo first \
    && echo second
COPY source /destination
