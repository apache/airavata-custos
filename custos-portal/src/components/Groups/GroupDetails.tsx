import React from "react";
import { Link, useParams } from "react-router-dom";
import { NavContainer } from "../NavContainer";
import { PageTitle } from "../PageTitle";
import { FaArrowLeft } from "react-icons/fa6";
import {
  Flex,
  Icon,
  Box,
  Text,
  Tabs, TabList, TabPanels, Tab, TabPanel
} from "@chakra-ui/react";
import { GroupSettings } from "./GroupSettings";
import { GroupMembers } from "./GroupMembers";
import { useApi } from "../../hooks/useApi";
import { BACKEND_URL } from "../../lib/constants";

interface CustomTabProps {
  children: React.ReactNode;
}

const CustomTab = ({ children }: CustomTabProps) => {
  return (
    <Tab
      _selected={{
        borderColor: 'default.default',
      }}
    >
      {children}
    </Tab>
  );
}

export const GroupDetails = () => {
  const { id, path } = useParams();
  const defaultIndex = path === 'members' ? 1 : 0;
  const basicGroupInfo = useApi(`${BACKEND_URL}/api/v1/group-management/groups/${id}`);

  return (
    <>
      <NavContainer activeTab="Groups">
        <Link to="/groups">
          <Flex alignItems='center' gap={2} color="default.secondary">
            <Icon as={FaArrowLeft}  />
            <Text fontWeight='bold' fontSize='sm'>Back to Groups</Text>
          </Flex>
        </Link>

        <Box mt={4}>
          <PageTitle>{basicGroupInfo?.data?.name}</PageTitle>
          <Text color="default.secondary" mt={2}>{basicGroupInfo?.data?.description}</Text>
        </Box>

        <Tabs 
          defaultIndex={defaultIndex}
          onChange={(index) => {
            window.history.replaceState(null, '', `/groups/${id}/${index === 0 ? 'settings' : 'members'}`);
          }}
          isLazy
        >
          <TabList mt={4}>
            <CustomTab>Settings</CustomTab>
            <CustomTab>Members</CustomTab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <GroupSettings groupId={id} />
            </TabPanel>
            <TabPanel>
              <GroupMembers groupId={id} />
            </TabPanel>
          </TabPanels>
      </Tabs>


      </NavContainer>
    </>
  );
}
