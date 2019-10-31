protoc -I/usr/local/include -I. \
    --go_out=plugins=grpc:. \
    --grpc-gateway_out=logtostderr=true,grpc_api_configuration=echo.yaml:. \
    --swagger_out=logtostderr=true:. \
    echo.proto
