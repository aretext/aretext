FROM base-image AS first_stage
RUN command_one \
    && command_two \
    && command_three

FROM runtime-image
COPY --from=first_stage /workspace/output /workspace/output
ENTRYPOINT ["/workspace/output"]
