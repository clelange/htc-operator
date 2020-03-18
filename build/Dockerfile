FROM centos

ENV OPERATOR=/usr/local/bin/htc-operator \
    USER_UID=1001 \
    USER_NAME=htc-controller

# install operator binary
COPY build/_output/bin/htc-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
