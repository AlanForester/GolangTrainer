# 1 choose a compiler OS
FROM golang:alpine AS builder
# 2 (optional) label the compiler image
LABEL stage=builder
# 3 (optional) install any compiler-only dependencies
RUN apk add --no-cache gcc libc-dev
WORKDIR /workspace
# 4 copy all the source files
COPY . .
# 5 build the GO program
RUN CGO_ENABLED=0 GOOS=linux go build -a
# 6 choose a runtime OS
FROM alpine AS final
# 7
ARG ENV
WORKDIR /
# 8 copy from builder the GO executable file
COPY --from=builder /workspace/slowly .
COPY --from=builder /workspace/_envs/env_$ENV.yaml ./_envs/
# 9 execute the program upon start
CMD [ "./slowly" ]