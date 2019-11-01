# Dockerfile
ARG EIS_VERSION
FROM ia_eisbase:$EIS_VERSION as eisbase

LABEL description="RestDataExport image"

WORKDIR /EIS/go/src/IEdgeInsights

FROM ia_common:$EIS_VERSION as common

FROM eisbase

COPY --from=common ${GO_WORK_DIR}/common/libs ${GO_WORK_DIR}/common/libs
COPY --from=common ${GO_WORK_DIR}/common/util ${GO_WORK_DIR}/common/util
COPY --from=common ${GO_WORK_DIR}/common/cmake ${GO_WORK_DIR}/common/cmake
COPY --from=common /usr/local/lib /usr/local/lib
COPY --from=common /usr/local/include /usr/local/include
COPY --from=common ${GO_WORK_DIR}/../EISMessageBus ${GO_WORK_DIR}/../EISMessageBus

ENV CGO_LDFLAGS "$CGO_LDFLAGS -leismsgbus -leismsgenv -leisutils"
ENV LD_LIBRARY_PATH ${LD_LIBRARY_PATH}:/usr/local/lib

COPY . ./RestDataExport/

RUN cd RestDataExport && go build RestDataExport.go
ENTRYPOINT ["./RestDataExport/RestDataExport"]
