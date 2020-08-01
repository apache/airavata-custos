# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import AgentProfileService_pb2 as AgentProfileService__pb2


class AgentProfileServiceStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.createAgent = channel.unary_unary(
        '/org.apache.custos.agent.profile.service.AgentProfileService/createAgent',
        request_serializer=AgentProfileService__pb2.AgentRequest.SerializeToString,
        response_deserializer=AgentProfileService__pb2.Agent.FromString,
        )
    self.updateAgent = channel.unary_unary(
        '/org.apache.custos.agent.profile.service.AgentProfileService/updateAgent',
        request_serializer=AgentProfileService__pb2.AgentRequest.SerializeToString,
        response_deserializer=AgentProfileService__pb2.Agent.FromString,
        )
    self.deleteAgent = channel.unary_unary(
        '/org.apache.custos.agent.profile.service.AgentProfileService/deleteAgent',
        request_serializer=AgentProfileService__pb2.AgentRequest.SerializeToString,
        response_deserializer=AgentProfileService__pb2.OperationStatus.FromString,
        )
    self.getAgent = channel.unary_unary(
        '/org.apache.custos.agent.profile.service.AgentProfileService/getAgent',
        request_serializer=AgentProfileService__pb2.AgentRequest.SerializeToString,
        response_deserializer=AgentProfileService__pb2.Agent.FromString,
        )


class AgentProfileServiceServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def createAgent(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def updateAgent(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def deleteAgent(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')

  def getAgent(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_AgentProfileServiceServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'createAgent': grpc.unary_unary_rpc_method_handler(
          servicer.createAgent,
          request_deserializer=AgentProfileService__pb2.AgentRequest.FromString,
          response_serializer=AgentProfileService__pb2.Agent.SerializeToString,
      ),
      'updateAgent': grpc.unary_unary_rpc_method_handler(
          servicer.updateAgent,
          request_deserializer=AgentProfileService__pb2.AgentRequest.FromString,
          response_serializer=AgentProfileService__pb2.Agent.SerializeToString,
      ),
      'deleteAgent': grpc.unary_unary_rpc_method_handler(
          servicer.deleteAgent,
          request_deserializer=AgentProfileService__pb2.AgentRequest.FromString,
          response_serializer=AgentProfileService__pb2.OperationStatus.SerializeToString,
      ),
      'getAgent': grpc.unary_unary_rpc_method_handler(
          servicer.getAgent,
          request_deserializer=AgentProfileService__pb2.AgentRequest.FromString,
          response_serializer=AgentProfileService__pb2.Agent.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'org.apache.custos.agent.profile.service.AgentProfileService', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
