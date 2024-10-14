FROM debian:bookworm-slim as devcontainer

RUN groupadd -g 10001 devcontainer && \
   useradd -u 10000 -g devcontainer devcontainer \
   && mkdir /home/devcontainer \
   && chown -R devcontainer:devcontainer /home/devcontainer \
   && mkdir /workspaces \
   && chown -R devcontainer:devcontainer /workspaces

RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y xz-utils net-tools vim curl wget git \
    && apt-get install -y gcc clang valgrind make scdoc cmake autogen autoconf flex bison \
    && apt-get clean autoclean \
    && apt-get autoremove --yes \
    && rm -rf /var/lib/{apt,dpkg,cache,log}/

WORKDIR /tmp

RUN curl "https://c9x.me/compile/release/qbe-1.2.tar.xz" -Lo qbe.tar.xz \
    && tar -xvf qbe.tar.xz
RUN cd qbe-1.2 \
    && make check \
    && make install \
    && rm -rf /tmp/qbe*

RUN git clone "https://git.sr.ht/~sircmpwn/harec"
RUN cd harec \
    && cp configs/linux.mk config.mk \
    && make \
    && make check \
    && make install \
    && rm -rf /tmp/harec

RUN git clone "https://git.sr.ht/~sircmpwn/hare" \
    && cd hare \
    && cp configs/linux.mk config.mk \
    && make \
    && make check \
    && make install \
    && rm -rf /tmp/hare

WORKDIR /home/devcontainer
