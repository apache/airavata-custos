# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import ClusterManagementService_pb2 as ClusterManagementService__pb2


class ClusterManagementServiceStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.getCustosServerCertificate = channel.unary_unary(
        '/org.apache.custos.cluster.management.service.ClusterManagementService/getCustosServerCertificate',
        request_serializer=ClusterManagementService__pb2.GetServerCertificateRequest.SerializeToString,
        response_deserializer=ClusterManagementService__pb2.GetServerCertificateResponse.FromString,
        )


class ClusterManagementServiceServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def getCustosServerCertificate(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_ClusterManagementServiceServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'getCustosServerCertificate': grpc.unary_unary_rpc_method_handler(
          servicer.getCustosServerCertificate,
          request_deserializer=ClusterManagementService__pb2.GetServerCertificateRequest.FromString,
          response_serializer=ClusterManagementService__pb2.GetServerCertificateResponse.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'org.apache.custos.cluster.management.service.ClusterManagementService', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
