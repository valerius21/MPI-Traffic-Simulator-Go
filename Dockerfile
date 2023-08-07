FROM golang:1.20.7-bookworm

# Update container
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    vim \
    tar \
    build-essential

# Install OpenMPI
RUN wget "https://download.open-mpi.org/release/open-mpi/v4.1/openmpi-4.1.5.tar.gz"
RUN tar -xvf openmpi-4.1.5.tar.gz
RUN cd openmpi-4.1.5 && ./configure --prefix=/usr/local && make && make install
RUN ldconfig
ENV OMPI_ALLOW_RUN_AS_ROOT=1
ENV OMPI_ALLOW_RUN_AS_ROOT_CONFIRM=1

WORKDIR /app
COPY . .

RUN git clone "https://github.com/sbromberger/gompi.git" && cd gompi
RUN go get "golang.org/x/tools/cmd/stringer"
RUN cd /app/gompi make install
RUN cd /app
