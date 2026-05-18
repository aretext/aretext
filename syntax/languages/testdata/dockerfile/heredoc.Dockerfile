FROM base-image
RUN <<EOF
line one
FROM text inside heredoc
# comment inside heredoc
EOF
COPY . /workspace
