onbuild add source /destination
OnBuild RUN echo "trigger"
shell ["/bin/sh", "-c"]
volume ["/data", "/cache"]
stopsignal SIGTERM
user app_user:app_group
