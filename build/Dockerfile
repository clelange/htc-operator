FROM golang:1.13-buster AS builder
ENV GO111MODULE=on
ARG CI_PROJECT_NAMESPACE=cms-cloud
ARG CI_PROJECT_NAME=htc-operator

# Copy the code from the host and compile it
WORKDIR ${GOPATH}/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE}/${CI_PROJECT_NAME}
COPY . ./
RUN go mod vendor
RUN CGO_ENABLED=1 GOOS=linux go build \
      -o /htc-operator \
      -ldflags '-extldflags "-static"' \
      -gcflags all=-trimpath=${GOPATH}/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE} \
      -asmflags all=-trimpath=${GOPATH}/src/gitlab.cern.ch/${CI_PROJECT_NAMESPACE} -mod=vendor \
      gitlab.cern.ch/${CI_PROJECT_NAMESPACE}/${CI_PROJECT_NAME}/cmd/manager


FROM gitlab-registry.cern.ch/clange/condorsubmit:latest

# install operator binary
COPY --from=builder /htc-operator ${OPERATOR}

COPY build/bin /usr/local/bin
