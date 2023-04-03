#Build .pb file
1. Install protoc plugin 
    
    resource : https://grpc.io/docs/languages/go/quickstart/
             : https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/grpc_json_transcoder_filter
2. Download   GOOGLE_APIS package
   
2. Generate  .pb file
   
     protoc -I<PATH_TO_GOOGLE_API>/googleapis  -I/airavata-data-lake/data-resource-management-service/drms-stubs/src/main/proto -I/airavata-data-lake/data-orchestrator/data-orchestrator-service/data-orchestrator-api-stub/src/main/proto  --include_imports --include_source_info     --descriptor_set_out=drms.pb  /airavata-data-lake/data-resource-management-service/drms-stubs/src/main/proto/preference/StoragePreferenceService.proto   /airavata-data-lake/data-resource-management-service/drms-stubs/src/main/proto/resource/DRMSResourceService.proto /airavata-data-lake/data-resource-management-service/drms-stubs/src/main/proto/storage/StorageService.proto /airavata-data-lake/data-orchestrator/data-orchestrator-service/data-orchestrator-api-stub/src/main/proto/notification.proto

3. copy .pb file into datalake-rest-proxy/src/main/resources

#  Run DRMS envoy proxy locally



1. Build docker image 
 `docker build -t drms-rest-proxy .`
   
2. Run docker image 
   `docker run --rm -it       -p 10000:50000        drms-rest-proxy:latest      -c /etc/envoy/envoy.yaml `
    