// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: wso2/discovery/subscription/subscription.proto

package org.wso2.choreo.connect.discovery.subscription;

public final class SubscriptionProto {
  private SubscriptionProto() {}
  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistryLite registry) {
  }

  public static void registerAllExtensions(
      com.google.protobuf.ExtensionRegistry registry) {
    registerAllExtensions(
        (com.google.protobuf.ExtensionRegistryLite) registry);
  }
  static final com.google.protobuf.Descriptors.Descriptor
    internal_static_wso2_discovery_subscription_Subscription_descriptor;
  static final 
    com.google.protobuf.GeneratedMessageV3.FieldAccessorTable
      internal_static_wso2_discovery_subscription_Subscription_fieldAccessorTable;

  public static com.google.protobuf.Descriptors.FileDescriptor
      getDescriptor() {
    return descriptor;
  }
  private static  com.google.protobuf.Descriptors.FileDescriptor
      descriptor;
  static {
    java.lang.String[] descriptorData = {
      "\n.wso2/discovery/subscription/subscripti" +
      "on.proto\022\033wso2.discovery.subscription\"\254\001" +
      "\n\014Subscription\022\026\n\016subscriptionId\030\001 \001(\t\022\020" +
      "\n\010policyId\030\002 \001(\t\022\r\n\005apiId\030\003 \001(\005\022\r\n\005appId" +
      "\030\004 \001(\005\022\031\n\021subscriptionState\030\005 \001(\t\022\021\n\ttim" +
      "eStamp\030\006 \001(\003\022\020\n\010tenantId\030\007 \001(\005\022\024\n\014tenant" +
      "Domain\030\010 \001(\tB\226\001\n.org.wso2.choreo.connect" +
      ".discovery.subscriptionB\021SubscriptionPro" +
      "toP\001ZOgithub.com/envoyproxy/go-control-p" +
      "lane/wso2/discovery/subscription;subscri" +
      "ptionb\006proto3"
    };
    descriptor = com.google.protobuf.Descriptors.FileDescriptor
      .internalBuildGeneratedFileFrom(descriptorData,
        new com.google.protobuf.Descriptors.FileDescriptor[] {
        });
    internal_static_wso2_discovery_subscription_Subscription_descriptor =
      getDescriptor().getMessageTypes().get(0);
    internal_static_wso2_discovery_subscription_Subscription_fieldAccessorTable = new
      com.google.protobuf.GeneratedMessageV3.FieldAccessorTable(
        internal_static_wso2_discovery_subscription_Subscription_descriptor,
        new java.lang.String[] { "SubscriptionId", "PolicyId", "ApiId", "AppId", "SubscriptionState", "TimeStamp", "TenantId", "TenantDomain", });
  }

  // @@protoc_insertion_point(outer_class_scope)
}