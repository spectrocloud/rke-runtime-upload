ARG KUBERNETES_VERSION
ARG CONTAINERD_VERSION
ARG CRICTL_VERSION
ARG RUNC_VERSION

FROM rancher/hardened-kubernetes:${KUBERNETES_VERSION} AS kubernetes
FROM rancher/hardened-containerd:${CONTAINERD_VERSION} AS containerd
FROM rancher/hardened-crictl:${CRICTL_VERSION} AS crictl
FROM rancher/hardened-runc:${RUNC_VERSION} AS runc

FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.
COPY containerd.service /containerd/etc/systemd/system/containerd.service

COPY --from=kubernetes \
    /usr/local/bin/kubectl \
    /usr/local/bin/kubelet \
    /k8s-runtime/usr/bin/

COPY --from=runc \
    /usr/local/bin/runc \
    /crictl/usr/bin/

COPY --from=crictl \
   /usr/local/bin/crictl \
   /crictl/usr/bin/

COPY --from=crictl \
   /usr/local/bin/crictl \
   /containerd/usr/local/bin/


COPY --from=runc \
    /usr/local/bin/runc \
    /containerd/usr/local/sbin/

COPY --from=containerd \
    /usr/local/bin/ctr \
    /crictl/usr/local/bin/

COPY --from=containerd \
    /usr/local/bin/containerd \
    /usr/local/bin/containerd-shim \
    /usr/local/bin/containerd-shim-runc-v1 \
    /usr/local/bin/containerd-shim-runc-v2 \
    /usr/local/bin/ctr \
    /containerd/usr/local/bin/

RUN cd /k8s-runtime && tar -cvzf k8s-runtime.tar.gz * && mv ./k8s-runtime.tar.gz /workspace/ && cd -
RUN cd /crictl && tar -cvzf crictl.tar.gz * && mv ./crictl.tar.gz /workspace/ && cd -
RUN cd /containerd && tar -cvzf containerd.tar.gz * && mv ./containerd.tar.gz /workspace/ && cd -

RUN ls -al

RUN ls -al ./

# Copy the go source
COPY . .


# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o k8s-runtime-uploader main.go

# Use distroless as minimal base image to package the manager binary
FROM gcr.io/distroless/static:latest
WORKDIR /

COPY --from=builder /workspace/k8s-runtime-uploader .
COPY --from=builder /workspace/k8s-runtime.tar.gz /k8s-runtime/k8s-runtime.tar.gz
COPY --from=builder /workspace/crictl.tar.gz /k8s-runtime/crictl.tar.gz
COPY --from=builder /workspace/containerd.tar.gz /k8s-runtime/containerd.tar.gz
ENTRYPOINT ["/k8s-runtime-uploader"]