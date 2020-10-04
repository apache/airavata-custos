import logging
import grpc

from custos.transport.settings import CustosServerClientSettings

from custos.server.integration.SharingManagementService_pb2_grpc import SharingManagementServiceStub

from custos.server.core.SharingService_pb2 import PermissionType, EntityType, Entity, SharingRequest, \
    PermissionTypeRequest, EntityRequest, EntityTypeRequest
from google.protobuf.json_format import MessageToJson
from custos.clients.utils.certificate_fetching_rest_client import CertificateFetchingRestClient

logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)


class SharingManagementClient(object):

    def __init__(self, custos_server_setting):
        self.custos_settings = custos_server_setting
        self.target = self.custos_settings.CUSTOS_SERVER_HOST + ":" + str(self.custos_settings.CUSTOS_SERVER_PORT)
        certManager = CertificateFetchingRestClient(custos_server_setting)
        certManager.load_certificate()
        with open(self.custos_settings.CUSTOS_CERT_PATH, 'rb') as f:
            trusted_certs = f.read()
        self.channel_credentials = grpc.ssl_channel_credentials(root_certificates=trusted_certs)
        self.channel = grpc.secure_channel(target=self.target, credentials=self.channel_credentials)
        self.sharing_mgt_client = SharingManagementServiceStub(self.channel)

    def create_entity_type(self, token, client_id, id, name, description):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            entity_type = EntityType(id=id, name=name, description=description)
            entity_type_req = EntityTypeRequest(client_id=client_id, entity_type=entity_type)
            return self.sharing_mgt_client.createEntityType(request=entity_type_req, metadata=metadata);
        except Exception:
            logger.exception("Error occurred while creating entity type with Id " + id)
            raise

    def create_permission_type(self, token, client_id, id, name, description):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            permission_type = PermissionType(id=id, name=name, description=description)
            permission_type_req = PermissionTypeRequest(client_id=client_id, permission_type=permission_type)
            return self.sharing_mgt_client.createPermissionType(request=permission_type_req, metadata=metadata);
        except Exception:
            logger.exception("Error occurred while creating permission entity type with Id " + id)
            raise

    def create_entity(self, token, client_id, id, name, description, owner_id, type, parent_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)
            entity = Entity(id=id, name=name, description=description, owner_id=owner_id, type=type,
                            parent_id=parent_id)
            entity_req = EntityRequest(client_id=client_id, entity=entity)
            return self.sharing_mgt_client.createEntity(request=entity_req, metadata=metadata);
        except Exception:
            logger.exception("Error occurred while creating  entity  with Id " + id)
            raise

    def share_entity_with_users(self, token, client_id, entity_id, permission_type, user_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            entity = Entity(id=entity_id)
            permission_type = PermissionType(id=permission_type)
            owner_ids = []
            owner_ids.append(user_id)
            cascade = True
            sharing_req = SharingRequest(client_id=client_id, entity=entity, permission_type=permission_type,
                                         owner_id=owner_ids, cascade=cascade)
            return self.sharing_mgt_client.shareEntityWithUsers(request=sharing_req, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while creating  entity  with Id " + entity_id)
            raise

    def share_entity_with_groups(self, token, client_id, entity_id, permission_type, group_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            entity = Entity(id=entity_id)
            permission_type = PermissionType(id=permission_type)
            owner_ids = []
            owner_ids.append(group_id)
            cascade = True
            sharing_req = SharingRequest(client_id=client_id, entity=entity, permission_type=permission_type,
                                         owner_id=owner_ids, cascade=cascade)
            return self.sharing_mgt_client.shareEntityWithGroups(request=sharing_req, metadata=metadata)
        except Exception:
            logger.exception("Error occurred while creating  entity  with Id " +
                             entity_id)
            raise

    def user_has_access(self, token, client_id, entity_id, permission_type, user_id):
        try:
            token = "Bearer " + token
            metadata = (('authorization', token),)

            entity = Entity(id=entity_id)
            permission_type = PermissionType(id=permission_type)
            owner_ids = []
            owner_ids.append(user_id)
            cascade = True
            sharing_req = SharingRequest(client_id=client_id, entity=entity, permission_type=permission_type,
                                         owner_id=owner_ids, cascade=cascade)
            resl = self.sharing_mgt_client.userHasAccess(request=sharing_req, metadata=metadata)

            if resl.status:
                return True
            else:
                return False

        except Exception:
            logger.exception("Error occurred while checking for permissions ")
            raise
