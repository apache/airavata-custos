# -*- coding: utf-8 -*-

#  Licensed to the Apache Software Foundation (ASF) under one or more
#  contributor license agreements.  See the NOTICE file distributed with
#  this work for additional information regarding copyright ownership.
#  The ASF licenses this file to You under the Apache License, Version 2.0
#  (the "License"); you may not use this file except in compliance with
#  the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: MessagingService.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='MessagingService.proto',
  package='org.apache.custos.messaging.service',
  syntax='proto3',
  serialized_options=b'P\001Z\004./pb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x16MessagingService.proto\x12#org.apache.custos.messaging.service\x1a\x1bgoogle/protobuf/empty.proto\"\x9a\x02\n\x07Message\x12\x14\n\x0c\x63reated_time\x18\x01 \x01(\x03\x12\x14\n\x0cservice_name\x18\x02 \x01(\t\x12\x12\n\nevent_type\x18\x03 \x01(\t\x12\x10\n\x08username\x18\x04 \x01(\t\x12\x11\n\tclient_id\x18\x05 \x01(\t\x12\x11\n\ttenant_id\x18\x06 \x01(\x03\x12P\n\nproperties\x18\x07 \x03(\x0b\x32<.org.apache.custos.messaging.service.Message.PropertiesEntry\x12\x12\n\nmessage_id\x18\x08 \x01(\t\x1a\x31\n\x0fPropertiesEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\">\n\x16MessageEnablingRequest\x12\x11\n\ttenant_id\x18\x01 \x01(\x03\x12\x11\n\tclient_id\x18\x02 \x01(\t\"\x18\n\x06Status\x12\x0e\n\x06status\x18\x01 \x01(\x08\"(\n\x17MessageEnablingResponse\x12\r\n\x05topic\x18\x01 \x01(\t2\xfe\x01\n\x10MessagingService\x12\x64\n\x07publish\x12,.org.apache.custos.messaging.service.Message\x1a+.org.apache.custos.messaging.service.Status\x12\x83\x01\n\x06\x65nable\x12;.org.apache.custos.messaging.service.MessageEnablingRequest\x1a<.org.apache.custos.messaging.service.MessageEnablingResponseB\x08P\x01Z\x04./pbb\x06proto3'
  ,
  dependencies=[google_dot_protobuf_dot_empty__pb2.DESCRIPTOR,])




_MESSAGE_PROPERTIESENTRY = _descriptor.Descriptor(
  name='PropertiesEntry',
  full_name='org.apache.custos.messaging.service.Message.PropertiesEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='org.apache.custos.messaging.service.Message.PropertiesEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='org.apache.custos.messaging.service.Message.PropertiesEntry.value', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=b'8\001',
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=326,
  serialized_end=375,
)

_MESSAGE = _descriptor.Descriptor(
  name='Message',
  full_name='org.apache.custos.messaging.service.Message',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='created_time', full_name='org.apache.custos.messaging.service.Message.created_time', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='service_name', full_name='org.apache.custos.messaging.service.Message.service_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='event_type', full_name='org.apache.custos.messaging.service.Message.event_type', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='username', full_name='org.apache.custos.messaging.service.Message.username', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='client_id', full_name='org.apache.custos.messaging.service.Message.client_id', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='tenant_id', full_name='org.apache.custos.messaging.service.Message.tenant_id', index=5,
      number=6, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='properties', full_name='org.apache.custos.messaging.service.Message.properties', index=6,
      number=7, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='message_id', full_name='org.apache.custos.messaging.service.Message.message_id', index=7,
      number=8, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_MESSAGE_PROPERTIESENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=93,
  serialized_end=375,
)


_MESSAGEENABLINGREQUEST = _descriptor.Descriptor(
  name='MessageEnablingRequest',
  full_name='org.apache.custos.messaging.service.MessageEnablingRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='tenant_id', full_name='org.apache.custos.messaging.service.MessageEnablingRequest.tenant_id', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='client_id', full_name='org.apache.custos.messaging.service.MessageEnablingRequest.client_id', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=377,
  serialized_end=439,
)


_STATUS = _descriptor.Descriptor(
  name='Status',
  full_name='org.apache.custos.messaging.service.Status',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='status', full_name='org.apache.custos.messaging.service.Status.status', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=441,
  serialized_end=465,
)


_MESSAGEENABLINGRESPONSE = _descriptor.Descriptor(
  name='MessageEnablingResponse',
  full_name='org.apache.custos.messaging.service.MessageEnablingResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='topic', full_name='org.apache.custos.messaging.service.MessageEnablingResponse.topic', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=467,
  serialized_end=507,
)

_MESSAGE_PROPERTIESENTRY.containing_type = _MESSAGE
_MESSAGE.fields_by_name['properties'].message_type = _MESSAGE_PROPERTIESENTRY
DESCRIPTOR.message_types_by_name['Message'] = _MESSAGE
DESCRIPTOR.message_types_by_name['MessageEnablingRequest'] = _MESSAGEENABLINGREQUEST
DESCRIPTOR.message_types_by_name['Status'] = _STATUS
DESCRIPTOR.message_types_by_name['MessageEnablingResponse'] = _MESSAGEENABLINGRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Message = _reflection.GeneratedProtocolMessageType('Message', (_message.Message,), {

  'PropertiesEntry' : _reflection.GeneratedProtocolMessageType('PropertiesEntry', (_message.Message,), {
    'DESCRIPTOR' : _MESSAGE_PROPERTIESENTRY,
    '__module__' : 'MessagingService_pb2'
    # @@protoc_insertion_point(class_scope:org.apache.custos.messaging.service.Message.PropertiesEntry)
    })
  ,
  'DESCRIPTOR' : _MESSAGE,
  '__module__' : 'MessagingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.messaging.service.Message)
  })
_sym_db.RegisterMessage(Message)
_sym_db.RegisterMessage(Message.PropertiesEntry)

MessageEnablingRequest = _reflection.GeneratedProtocolMessageType('MessageEnablingRequest', (_message.Message,), {
  'DESCRIPTOR' : _MESSAGEENABLINGREQUEST,
  '__module__' : 'MessagingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.messaging.service.MessageEnablingRequest)
  })
_sym_db.RegisterMessage(MessageEnablingRequest)

Status = _reflection.GeneratedProtocolMessageType('Status', (_message.Message,), {
  'DESCRIPTOR' : _STATUS,
  '__module__' : 'MessagingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.messaging.service.Status)
  })
_sym_db.RegisterMessage(Status)

MessageEnablingResponse = _reflection.GeneratedProtocolMessageType('MessageEnablingResponse', (_message.Message,), {
  'DESCRIPTOR' : _MESSAGEENABLINGRESPONSE,
  '__module__' : 'MessagingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.messaging.service.MessageEnablingResponse)
  })
_sym_db.RegisterMessage(MessageEnablingResponse)


DESCRIPTOR._options = None
_MESSAGE_PROPERTIESENTRY._options = None

_MESSAGINGSERVICE = _descriptor.ServiceDescriptor(
  name='MessagingService',
  full_name='org.apache.custos.messaging.service.MessagingService',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=510,
  serialized_end=764,
  methods=[
  _descriptor.MethodDescriptor(
    name='publish',
    full_name='org.apache.custos.messaging.service.MessagingService.publish',
    index=0,
    containing_service=None,
    input_type=_MESSAGE,
    output_type=_STATUS,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='enable',
    full_name='org.apache.custos.messaging.service.MessagingService.enable',
    index=1,
    containing_service=None,
    input_type=_MESSAGEENABLINGREQUEST,
    output_type=_MESSAGEENABLINGRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_MESSAGINGSERVICE)

DESCRIPTOR.services_by_name['MessagingService'] = _MESSAGINGSERVICE

# @@protoc_insertion_point(module_scope)
