FROM golang:1.13-buster AS builder
ENV GO111MODULE=on
ARG CI_PROJECT_NAMESPACE
ARG CI_PROJECT_NAME

# Copy the code from the host and compile it
WORKDIR $GOPATH/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE}/${CI_PROJECT_NAME}
COPY . ./
RUN go mod vendor
RUN echo ${CI_PROJECT_NAMESPACE} ${CI_PROJECT_NAME}
RUN CGO_ENABLED=0 GOOS=linux go build \
      -o /htc-operator \
      -gcflags all=-trimpath=/root/go/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE} \
      -asmflags all=-trimpath=/root/go/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE} -mod=vendor \
      gitlab.cern.ch/${CI_PROJECT_NAMESPACE}/${CI_PROJECT_NAME}/cmd/manager

FROM gitlab-registry.cern.ch/clange/condorsubmit:latest

ARG USER_NAME=tbareiki
ENV OPERATOR=/usr/local/bin/htc-operator
ENV USER_UID=1001

RUN python -m pip --no-cache-dir install s3cmd

# install operator binary
COPY /htc-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
