/* Generated by the protocol buffer compiler.  DO NOT EDIT! */
/* Generated from: api/core/address.proto */

/* Do not generate deprecated warnings for self */
#ifndef PROTOBUF_C__NO_DEPRECATED
#define PROTOBUF_C__NO_DEPRECATED
#endif

#include "core/address.pb-c.h"
void   core__socket_address__init
                     (Core__SocketAddress         *message)
{
  static const Core__SocketAddress init_value = CORE__SOCKET_ADDRESS__INIT;
  *message = init_value;
}
size_t core__socket_address__get_packed_size
                     (const Core__SocketAddress *message)
{
  assert(message->base.descriptor == &core__socket_address__descriptor);
  return protobuf_c_message_get_packed_size ((const ProtobufCMessage*)(message));
}
size_t core__socket_address__pack
                     (const Core__SocketAddress *message,
                      uint8_t       *out)
{
  assert(message->base.descriptor == &core__socket_address__descriptor);
  return protobuf_c_message_pack ((const ProtobufCMessage*)message, out);
}
size_t core__socket_address__pack_to_buffer
                     (const Core__SocketAddress *message,
                      ProtobufCBuffer *buffer)
{
  assert(message->base.descriptor == &core__socket_address__descriptor);
  return protobuf_c_message_pack_to_buffer ((const ProtobufCMessage*)message, buffer);
}
Core__SocketAddress *
       core__socket_address__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data)
{
  return (Core__SocketAddress *)
     protobuf_c_message_unpack (&core__socket_address__descriptor,
                                allocator, len, data);
}
void   core__socket_address__free_unpacked
                     (Core__SocketAddress *message,
                      ProtobufCAllocator *allocator)
{
  if(!message)
    return;
  assert(message->base.descriptor == &core__socket_address__descriptor);
  protobuf_c_message_free_unpacked ((ProtobufCMessage*)message, allocator);
}
void   core__cidr_range__init
                     (Core__CidrRange         *message)
{
  static const Core__CidrRange init_value = CORE__CIDR_RANGE__INIT;
  *message = init_value;
}
size_t core__cidr_range__get_packed_size
                     (const Core__CidrRange *message)
{
  assert(message->base.descriptor == &core__cidr_range__descriptor);
  return protobuf_c_message_get_packed_size ((const ProtobufCMessage*)(message));
}
size_t core__cidr_range__pack
                     (const Core__CidrRange *message,
                      uint8_t       *out)
{
  assert(message->base.descriptor == &core__cidr_range__descriptor);
  return protobuf_c_message_pack ((const ProtobufCMessage*)message, out);
}
size_t core__cidr_range__pack_to_buffer
                     (const Core__CidrRange *message,
                      ProtobufCBuffer *buffer)
{
  assert(message->base.descriptor == &core__cidr_range__descriptor);
  return protobuf_c_message_pack_to_buffer ((const ProtobufCMessage*)message, buffer);
}
Core__CidrRange *
       core__cidr_range__unpack
                     (ProtobufCAllocator  *allocator,
                      size_t               len,
                      const uint8_t       *data)
{
  return (Core__CidrRange *)
     protobuf_c_message_unpack (&core__cidr_range__descriptor,
                                allocator, len, data);
}
void   core__cidr_range__free_unpacked
                     (Core__CidrRange *message,
                      ProtobufCAllocator *allocator)
{
  if(!message)
    return;
  assert(message->base.descriptor == &core__cidr_range__descriptor);
  protobuf_c_message_free_unpacked ((ProtobufCMessage*)message, allocator);
}
static const ProtobufCEnumValue core__socket_address__protocol__enum_values_by_number[2] =
{
  { "TCP", "CORE__SOCKET_ADDRESS__PROTOCOL__TCP", 0 },
  { "UDP", "CORE__SOCKET_ADDRESS__PROTOCOL__UDP", 1 },
};
static const ProtobufCIntRange core__socket_address__protocol__value_ranges[] = {
{0, 0},{0, 2}
};
static const ProtobufCEnumValueIndex core__socket_address__protocol__enum_values_by_name[2] =
{
  { "TCP", 0 },
  { "UDP", 1 },
};
const ProtobufCEnumDescriptor core__socket_address__protocol__descriptor =
{
  PROTOBUF_C__ENUM_DESCRIPTOR_MAGIC,
  "core.SocketAddress.Protocol",
  "Protocol",
  "Core__SocketAddress__Protocol",
  "core",
  2,
  core__socket_address__protocol__enum_values_by_number,
  2,
  core__socket_address__protocol__enum_values_by_name,
  1,
  core__socket_address__protocol__value_ranges,
  NULL,NULL,NULL,NULL   /* reserved[1234] */
};
static const ProtobufCFieldDescriptor core__socket_address__field_descriptors[3] =
{
  {
    "protocol",
    1,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_ENUM,
    0,   /* quantifier_offset */
    offsetof(Core__SocketAddress, protocol),
    &core__socket_address__protocol__descriptor,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
  {
    "port",
    2,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_UINT32,
    0,   /* quantifier_offset */
    offsetof(Core__SocketAddress, port),
    NULL,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
  {
    "ipv4",
    3,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_UINT32,
    0,   /* quantifier_offset */
    offsetof(Core__SocketAddress, ipv4),
    NULL,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
};
static const unsigned core__socket_address__field_indices_by_name[] = {
  2,   /* field[2] = ipv4 */
  1,   /* field[1] = port */
  0,   /* field[0] = protocol */
};
static const ProtobufCIntRange core__socket_address__number_ranges[1 + 1] =
{
  { 1, 0 },
  { 0, 3 }
};
const ProtobufCMessageDescriptor core__socket_address__descriptor =
{
  PROTOBUF_C__MESSAGE_DESCRIPTOR_MAGIC,
  "core.SocketAddress",
  "SocketAddress",
  "Core__SocketAddress",
  "core",
  sizeof(Core__SocketAddress),
  3,
  core__socket_address__field_descriptors,
  core__socket_address__field_indices_by_name,
  1,  core__socket_address__number_ranges,
  (ProtobufCMessageInit) core__socket_address__init,
  NULL,NULL,NULL    /* reserved[123] */
};
static const ProtobufCFieldDescriptor core__cidr_range__field_descriptors[2] =
{
  {
    "address_prefix",
    1,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_STRING,
    0,   /* quantifier_offset */
    offsetof(Core__CidrRange, address_prefix),
    NULL,
    &protobuf_c_empty_string,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
  {
    "prefix_len",
    2,
    PROTOBUF_C_LABEL_NONE,
    PROTOBUF_C_TYPE_UINT32,
    0,   /* quantifier_offset */
    offsetof(Core__CidrRange, prefix_len),
    NULL,
    NULL,
    0,             /* flags */
    0,NULL,NULL    /* reserved1,reserved2, etc */
  },
};
static const unsigned core__cidr_range__field_indices_by_name[] = {
  0,   /* field[0] = address_prefix */
  1,   /* field[1] = prefix_len */
};
static const ProtobufCIntRange core__cidr_range__number_ranges[1 + 1] =
{
  { 1, 0 },
  { 0, 2 }
};
const ProtobufCMessageDescriptor core__cidr_range__descriptor =
{
  PROTOBUF_C__MESSAGE_DESCRIPTOR_MAGIC,
  "core.CidrRange",
  "CidrRange",
  "Core__CidrRange",
  "core",
  sizeof(Core__CidrRange),
  2,
  core__cidr_range__field_descriptors,
  core__cidr_range__field_indices_by_name,
  1,  core__cidr_range__number_ranges,
  (ProtobufCMessageInit) core__cidr_range__init,
  NULL,NULL,NULL    /* reserved[123] */
};
