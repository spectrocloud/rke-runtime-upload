# rke-runtime-upload
Upload runtime binaries to s3 bucket

Run below command to build and run docker images

examples: 

```docker build --platform linux/amd64 --build-arg KUBERNETES_VERSION=v1.24.3-rke2r1-build20220713 --build-arg CONTAINERD_VERSION=v1.6.6-k3s1-build20220606 --build-arg CRICTL_VERSION=v1.24.0-build20220506 --build-arg RUNC_VERSION=v1.1.2-build20220606 -t ImageName:Tag .```

```docker run --env AWS_ACCESS_KEY_ID=<Access Key> --env AWS_SECRET_ACCESS_KEY=<Secret Key> --env AWS_BUCKET_NAME=<Bucket> --env AWS_REGION=us-east-1 --env KUBERNETES_VERSION=1.24.3-rke2r1-build20220713  imageName:Tag```
