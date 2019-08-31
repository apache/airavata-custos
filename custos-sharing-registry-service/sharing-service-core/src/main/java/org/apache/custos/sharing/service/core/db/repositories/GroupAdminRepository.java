package org.apache.custos.sharing.service.core.db.repositories;

import org.apache.custos.sharing.service.core.models.GroupAdmin;
import org.apache.custos.sharing.service.core.db.entities.GroupAdminEntity;
import org.apache.custos.sharing.service.core.db.entities.GroupAdminPK;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GroupAdminRepository extends AbstractRepository<GroupAdmin, GroupAdminEntity, GroupAdminPK> {

    private final static Logger logger = LoggerFactory.getLogger(GroupAdminRepository.class);

    public GroupAdminRepository() {
        super(GroupAdmin.class, GroupAdminEntity.class);
    }

}
