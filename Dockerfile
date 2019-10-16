# Dockerfile
ARG EIS_VERSION
FROM ia_eisbase:$EIS_VERSION as eisbase

LABEL description="RestDataExport image"

WORKDIR /EIS/go/src/IEdgeInsights

FROM ia_common:$EIS_VERSION as common

FROM eisbase

COPY --from=common /libs ${GO_WORK_DIR}/libs
COPY --from=common /util ${GO_WORK_DIR}/util

RUN cd ${GO_WORK_DIR}/libs/EISMessageBus && \
    rm -rf build deps && mkdir -p build && cd build && \
    cmake -DWITH_GO=ON .. && \
    make && \
    make install

ENV MSGBUS_DIR $GO_WORK_DIR/libs/EISMessageBus
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:$MSGBUS_DIR/build/
ENV PKG_CONFIG_PATH $PKG_CONFIG_PATH:$MSGBUS_DIR/build/
ENV CGO_CFLAGS -I$MSGBUS_DIR/include/
ENV CGO_LDFLAGS -L$MSGBUS_DIR/build -leismsgbus
ENV LD_LIBRARY_PATH ${LD_LIBRARY_PATH}:/usr/local/lib

RUN ln -s ${GO_WORK_DIR}/libs/EISMessageBus/go/EISMessageBus/ $GOPATH/src/EISMessageBus

COPY . ./RestDataExport/

RUN cd RestDataExport && go build RestDataExport.go
ENTRYPOINT ["./RestDataExport/RestDataExport"]
