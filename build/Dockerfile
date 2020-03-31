FROM gitlab-registry.cern.ch/clange/condorsubmit:latest

ARG USER_NAME=tbareiki
ENV OPERATOR=/usr/local/bin/htc-operator
ENV USER_UID=1001

RUN python -m pip --no-cache-dir install s3cmd

# install operator binary
COPY build/_output/bin/htc-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
