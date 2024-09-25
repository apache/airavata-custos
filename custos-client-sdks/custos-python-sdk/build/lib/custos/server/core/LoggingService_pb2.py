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
# source: LoggingService.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='LoggingService.proto',
  package='org.apache.custos.logging.service',
  syntax='proto3',
  serialized_options=b'P\001Z\004./pb',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x14LoggingService.proto\x12!org.apache.custos.logging.service\x1a\x1bgoogle/protobuf/empty.proto\"\x97\x01\n\x08LogEvent\x12\x14\n\x0c\x63reated_time\x18\x01 \x01(\x03\x12\x14\n\x0cservice_name\x18\x02 \x01(\t\x12\x12\n\nevent_type\x18\x03 \x01(\t\x12\x10\n\x08username\x18\x04 \x01(\t\x12\x11\n\tclient_id\x18\x05 \x01(\t\x12\x11\n\ttenant_id\x18\x06 \x01(\x03\x12\x13\n\x0b\x65xternal_ip\x18\x07 \x01(\t\"\x18\n\x06Status\x12\x0e\n\x06status\x18\x01 \x01(\x08\"\xcb\x01\n\x0fLogEventRequest\x12\x11\n\ttenant_id\x18\x01 \x01(\x03\x12\x12\n\nstart_time\x18\x02 \x01(\x03\x12\x10\n\x08\x65nd_time\x18\x03 \x01(\x03\x12\x11\n\tclient_id\x18\x04 \x01(\t\x12\x10\n\x08username\x18\x05 \x01(\t\x12\x11\n\tremote_ip\x18\x06 \x01(\t\x12\x14\n\x0cservice_name\x18\x07 \x01(\t\x12\x12\n\nevent_type\x18\x08 \x01(\t\x12\x0e\n\x06offset\x18\t \x01(\x05\x12\r\n\x05limit\x18\n \x01(\x05\"H\n\tLogEvents\x12;\n\x06\x65vents\x18\x01 \x03(\x0b\x32+.org.apache.custos.logging.service.LogEvent\"C\n\x1bLoggingConfigurationRequest\x12\x11\n\ttenant_id\x18\x01 \x01(\x03\x12\x11\n\tclient_id\x18\x02 \x01(\t2\xd9\x03\n\x0eLoggingService\x12\x65\n\x0b\x61\x64\x64LogEvent\x12+.org.apache.custos.logging.service.LogEvent\x1a).org.apache.custos.logging.service.Status\x12p\n\x0cgetLogEvents\x12\x32.org.apache.custos.logging.service.LogEventRequest\x1a,.org.apache.custos.logging.service.LogEvents\x12y\n\x0cisLogEnabled\x12>.org.apache.custos.logging.service.LoggingConfigurationRequest\x1a).org.apache.custos.logging.service.Status\x12s\n\x06\x65nable\x12>.org.apache.custos.logging.service.LoggingConfigurationRequest\x1a).org.apache.custos.logging.service.StatusB\x08P\x01Z\x04./pbb\x06proto3'
  ,
  dependencies=[google_dot_protobuf_dot_empty__pb2.DESCRIPTOR,])




_LOGEVENT = _descriptor.Descriptor(
  name='LogEvent',
  full_name='org.apache.custos.logging.service.LogEvent',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='created_time', full_name='org.apache.custos.logging.service.LogEvent.created_time', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='service_name', full_name='org.apache.custos.logging.service.LogEvent.service_name', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='event_type', full_name='org.apache.custos.logging.service.LogEvent.event_type', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='username', full_name='org.apache.custos.logging.service.LogEvent.username', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='client_id', full_name='org.apache.custos.logging.service.LogEvent.client_id', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='tenant_id', full_name='org.apache.custos.logging.service.LogEvent.tenant_id', index=5,
      number=6, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='external_ip', full_name='org.apache.custos.logging.service.LogEvent.external_ip', index=6,
      number=7, type=9, cpp_type=9, label=1,
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
  serialized_start=89,
  serialized_end=240,
)


_STATUS = _descriptor.Descriptor(
  name='Status',
  full_name='org.apache.custos.logging.service.Status',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='status', full_name='org.apache.custos.logging.service.Status.status', index=0,
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
  serialized_start=242,
  serialized_end=266,
)


_LOGEVENTREQUEST = _descriptor.Descriptor(
  name='LogEventRequest',
  full_name='org.apache.custos.logging.service.LogEventRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='tenant_id', full_name='org.apache.custos.logging.service.LogEventRequest.tenant_id', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='start_time', full_name='org.apache.custos.logging.service.LogEventRequest.start_time', index=1,
      number=2, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='end_time', full_name='org.apache.custos.logging.service.LogEventRequest.end_time', index=2,
      number=3, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='client_id', full_name='org.apache.custos.logging.service.LogEventRequest.client_id', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='username', full_name='org.apache.custos.logging.service.LogEventRequest.username', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='remote_ip', full_name='org.apache.custos.logging.service.LogEventRequest.remote_ip', index=5,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='service_name', full_name='org.apache.custos.logging.service.LogEventRequest.service_name', index=6,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='event_type', full_name='org.apache.custos.logging.service.LogEventRequest.event_type', index=7,
      number=8, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='offset', full_name='org.apache.custos.logging.service.LogEventRequest.offset', index=8,
      number=9, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='limit', full_name='org.apache.custos.logging.service.LogEventRequest.limit', index=9,
      number=10, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
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
  serialized_start=269,
  serialized_end=472,
)


_LOGEVENTS = _descriptor.Descriptor(
  name='LogEvents',
  full_name='org.apache.custos.logging.service.LogEvents',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='events', full_name='org.apache.custos.logging.service.LogEvents.events', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
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
  serialized_start=474,
  serialized_end=546,
)


_LOGGINGCONFIGURATIONREQUEST = _descriptor.Descriptor(
  name='LoggingConfigurationRequest',
  full_name='org.apache.custos.logging.service.LoggingConfigurationRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='tenant_id', full_name='org.apache.custos.logging.service.LoggingConfigurationRequest.tenant_id', index=0,
      number=1, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='client_id', full_name='org.apache.custos.logging.service.LoggingConfigurationRequest.client_id', index=1,
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
  serialized_start=548,
  serialized_end=615,
)

_LOGEVENTS.fields_by_name['events'].message_type = _LOGEVENT
DESCRIPTOR.message_types_by_name['LogEvent'] = _LOGEVENT
DESCRIPTOR.message_types_by_name['Status'] = _STATUS
DESCRIPTOR.message_types_by_name['LogEventRequest'] = _LOGEVENTREQUEST
DESCRIPTOR.message_types_by_name['LogEvents'] = _LOGEVENTS
DESCRIPTOR.message_types_by_name['LoggingConfigurationRequest'] = _LOGGINGCONFIGURATIONREQUEST
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

LogEvent = _reflection.GeneratedProtocolMessageType('LogEvent', (_message.Message,), {
  'DESCRIPTOR' : _LOGEVENT,
  '__module__' : 'LoggingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.logging.service.LogEvent)
  })
_sym_db.RegisterMessage(LogEvent)

Status = _reflection.GeneratedProtocolMessageType('Status', (_message.Message,), {
  'DESCRIPTOR' : _STATUS,
  '__module__' : 'LoggingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.logging.service.Status)
  })
_sym_db.RegisterMessage(Status)

LogEventRequest = _reflection.GeneratedProtocolMessageType('LogEventRequest', (_message.Message,), {
  'DESCRIPTOR' : _LOGEVENTREQUEST,
  '__module__' : 'LoggingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.logging.service.LogEventRequest)
  })
_sym_db.RegisterMessage(LogEventRequest)

LogEvents = _reflection.GeneratedProtocolMessageType('LogEvents', (_message.Message,), {
  'DESCRIPTOR' : _LOGEVENTS,
  '__module__' : 'LoggingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.logging.service.LogEvents)
  })
_sym_db.RegisterMessage(LogEvents)

LoggingConfigurationRequest = _reflection.GeneratedProtocolMessageType('LoggingConfigurationRequest', (_message.Message,), {
  'DESCRIPTOR' : _LOGGINGCONFIGURATIONREQUEST,
  '__module__' : 'LoggingService_pb2'
  # @@protoc_insertion_point(class_scope:org.apache.custos.logging.service.LoggingConfigurationRequest)
  })
_sym_db.RegisterMessage(LoggingConfigurationRequest)


DESCRIPTOR._options = None

_LOGGINGSERVICE = _descriptor.ServiceDescriptor(
  name='LoggingService',
  full_name='org.apache.custos.logging.service.LoggingService',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=618,
  serialized_end=1091,
  methods=[
  _descriptor.MethodDescriptor(
    name='addLogEvent',
    full_name='org.apache.custos.logging.service.LoggingService.addLogEvent',
    index=0,
    containing_service=None,
    input_type=_LOGEVENT,
    output_type=_STATUS,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='getLogEvents',
    full_name='org.apache.custos.logging.service.LoggingService.getLogEvents',
    index=1,
    containing_service=None,
    input_type=_LOGEVENTREQUEST,
    output_type=_LOGEVENTS,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='isLogEnabled',
    full_name='org.apache.custos.logging.service.LoggingService.isLogEnabled',
    index=2,
    containing_service=None,
    input_type=_LOGGINGCONFIGURATIONREQUEST,
    output_type=_STATUS,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='enable',
    full_name='org.apache.custos.logging.service.LoggingService.enable',
    index=3,
    containing_service=None,
    input_type=_LOGGINGCONFIGURATIONREQUEST,
    output_type=_STATUS,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_LOGGINGSERVICE)

DESCRIPTOR.services_by_name['LoggingService'] = _LOGGINGSERVICE

# @@protoc_insertion_point(module_scope)
