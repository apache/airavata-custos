import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  VStack,
  FormControl,
  FormLabel,
  Input,
  Textarea,
  Box,
  useToast
} from '@chakra-ui/react';
import { Group } from '../../interfaces/Groups';
import { useState } from 'react';
import { TagsInput } from "react-tag-input-component";
import axios from 'axios';
import { BACKEND_URL } from '../../lib/constants';
import { useNavigate } from 'react-router-dom';

export const CreateGroupModal = ({isOpen, onClose, auth }: {
  isOpen: boolean;
  onClose: () => void;
  auth: any;
}) => {
  const [group, setGroup] = useState({} as Group);
  const toast = useToast();
  const navigate = useNavigate();

  const handleSubmit = () => {
    if (!group.id || !group.name || !group.description) {
      toast({
        title: 'Invalid input',
        description: 'Please fill out all required fields',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const groupObj = {
      ...group,
      owner_id: auth?.user?.profile?.email,
    }

    axios.post(`${BACKEND_URL}/api/v1/group-management/groups`, groupObj, {
      headers: {
        Authorization: `Bearer ${auth.user.access_token}`
      }
    }).then(() => {
      onClose();
      navigate(0);
    }).catch((err) => {
      console.log(err);
      toast({
        title: 'An error occurred',
        description: err.response?.data?.message || 'Please try again later',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    });
  };

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Create Group</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack
              spacing={4}
            >
              <FormControl>
                <FormLabel>Id</FormLabel>
                <Input placeholder='must be unique from all other groups'
                  value={group.id}
                  onChange={(e) => setGroup((prev) => ({...prev, id: e.target.value}))}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Name</FormLabel>
                <Input placeholder='i.e. Dashboard Viewers'
                  value={group.name}
                  onChange={(e) => setGroup((prev) => ({...prev, name: e.target.value}))}
                />
              </FormControl>


              <FormControl>
                <FormLabel>Description</FormLabel>
                <Textarea placeholder='i.e. viewers and editors of the dashboard'
                  value={group.description}
                  onChange={(e) => setGroup((prev) => ({...prev, description: e.target.value}))}
                />
              </FormControl>

              <FormControl>
                <FormLabel>Parent Group Id (optional)</FormLabel>
                <Input 
                  placeholder=''
                  value={group.parent_id}
                  onChange={(e) => setGroup((prev) => ({...prev, parent_id: e.target.value}))}
                />
              </FormControl>


              <FormControl>
                <FormLabel>Attributes (optional)</FormLabel>
                {
                  group.attributes?.map((attr) => {
                    return (
                      <Box key={attr.id} border='1px solid lightgray' rounded='md' mt={2} p={2}>
                        <Input placeholder='key'
                          value={attr.key}
                          onChange={(e) => {
                            setGroup((prev) => ({
                              ...prev,
                              attributes: prev.attributes?.map((a) => {
                                if (a.id === attr.id) {
                                  return {...a, key: e.target.value}
                                }
                                return a;
                              })
                            }))
                          }}
                        />

                        <TagsInput
                          value={attr.value}
                          onChange={(tags) => {
                            setGroup((prev) => ({
                              ...prev,
                              attributes: prev.attributes?.map((a) => {
                                if (a.id === attr.id) {
                                  return {...a, value: tags}
                                }
                                return a;
                              })
                            }))
                          }}
                          name="fruits"
                          placeHolder="i.e. grafana:viewer"
                        />

                        <Button
                          size='sm'
                          colorScheme='red'
                          mt={2}
                          onClick={() => {
                            setGroup((prev) => ({
                              ...prev,
                              attributes: prev.attributes?.filter((a) => a.id !== attr.id)
                            }))
                          }}
                        >
                          Remove
                        </Button>
                      </Box>
                    )
                  })
                }

                <Button
                  mt={2}
                  onClick={() => {
                    setGroup((prev) => ({
                      ...prev,
                      attributes: [
                        ...(prev.attributes || []),
                        {
                          id: Date.now(),
                          key: '',
                          value: []
                        }
                      ]
                    }))
                  }}
                >
                  Add Attribute
                </Button>
              </FormControl>

              <FormControl>
                <FormLabel>Client roles (enter to create a new role)</FormLabel>
                <TagsInput
                  value={group.client_roles}
                  onChange={(tags) => {
                    setGroup((prev) => ({...prev, client_roles: tags}))}}
                  name="fruits"
                  placeHolder="i.e. grafana:viewer"
                />
              </FormControl>


              <FormControl>
                <FormLabel>Realm roles (enter to create a new role)</FormLabel>
                <TagsInput
                  value={group.realm_roles}
                  onChange={(tags) => {
                    setGroup((prev) => ({...prev, realm_roles: tags}))}}
                  name="fruits"
                  placeHolder="i.e. grafana:viewer"
                />
              </FormControl>

              <Button colorScheme='blue'  onClick={handleSubmit}>
                Submit
              </Button>
            </VStack>

            
          </ModalBody>

          <ModalFooter>
           
          </ModalFooter>
        </ModalContent>

      </Modal>
    </>
  )
}