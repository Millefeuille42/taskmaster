FROM debian:bookworm-slim AS devcontainer

RUN groupadd -g 10001 devcontainer && \
   useradd -u 10000 -g devcontainer devcontainer \
   && mkdir /home/devcontainer \
   && chown -R devcontainer:devcontainer /home/devcontainer \
   && mkdir /workspaces \
   && chown -R devcontainer:devcontainer /workspaces

RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y xz-utils net-tools vim curl wget git \
    && apt-get install -y gcc valgrind make scdoc cmake autogen autoconf flex bison \
    && apt-get clean autoclean \
    && apt-get autoremove --yes \
    && rm -rf /var/lib/{apt,dpkg,cache,log}/

WORKDIR /tmp

RUN curl "https://c9x.me/compile/release/qbe-1.2.tar.xz" -Lo qbe.tar.xz \
    && tar -xvf qbe.tar.xz \
    && cd qbe-1.2 \
    && make -j$(nproc) \
    && make -j$(nproc) check \
    && make -j$(nproc) install \
    && rm -rf /tmp/qbe*

RUN curl "https://git.sr.ht/~sircmpwn/harec/archive/0.24.2.tar.gz" -Lo harec.tar.gz \
    && tar -xvf harec.tar.gz \
    && cd harec-0.24.2 \
    && cp configs/linux.mk config.mk \
    && make -j$(nproc) \
    && make -j$(nproc) check \
    && make -j$(nproc) install \
    && rm -rf /tmp/harec

RUN curl "https://git.sr.ht/~sircmpwn/hare/archive/0.24.2.tar.gz" -Lo hare.tar.gz \
    && tar -xvf hare.tar.gz \
    && cd hare-0.24.2 \
    && cp configs/linux.mk config.mk \
    && make -j$(nproc) \
    && make -j$(nproc) check \
    && make -j$(nproc) install \
    && rm -rf /tmp/hare

WORKDIR /home/devcontainer
