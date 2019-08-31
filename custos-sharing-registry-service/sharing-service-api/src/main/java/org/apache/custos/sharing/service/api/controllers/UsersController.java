package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.User;
import org.apache.custos.sharing.service.core.exceptions.ResourceNotFoundException;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/user")
public class UsersController {

    @Autowired
    private static SharingRegistryService sharingRegistryService;

    /**
     <p>API method to register a user in the system</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<String> createUser(@RequestBody User user){
        String userId = sharingRegistryService.createUser(user);
        if(userId.equals(user.getUserId())){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to create the user");
        }
    }

    /**
     <p>API method to update existing user</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateUser(@RequestBody User user){
        boolean response = sharingRegistryService.updatedUser(user);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to update the user");
        }
    }

    /**
     <p>API method to check User Exists</p>
     */
    @RequestMapping(value = "/exists/domainId/{domainId}/userId/{userId}", method = RequestMethod.GET)
    public @ResponseBody
    Map<String, Boolean> isUserExists(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId){
        boolean response = sharingRegistryService.isUserExists(domainId, userId);
        return Collections.singletonMap("exists", response);
    }

    /**
     <p>API method to delete user</p>
     */
    @RequestMapping(value="/id/{userId}/domain/{domainId}",method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteUser(@PathVariable("userId") String userId, @PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deleteUser(domainId, userId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            throw new SharingRegistryException("Failed to delete the user");
        }
    }

    /**
     <p>API method to get a user</p>
     */
    @RequestMapping(value = "/domain/{domainId}/user/{userId}", method = RequestMethod.GET)
    public User getUser(@PathVariable("domainId") String domainId,@PathVariable("userId") String userId){
        User user = sharingRegistryService.getUser(domainId, userId);
        if(user != null){
            return user;
        }else {
            throw new ResourceNotFoundException("Could not find the user with userId: "+ userId + " for domainId: "+ domainId);
        }
    }

    /**
     <p>API method to get a list of users in a specific domain.</p>
     <li>domainId : Domain id</li>
     <li>offset : Starting result number</li>
     <li>limit : Number of max results to be sent</li>
     */
    @RequestMapping(value = "/domainId/{domainId}/offset/{offset}/limit/{limit}")
    public List<User> getUsers(@PathVariable("domainId") String domainId, @PathVariable("offset") int offset, @PathVariable("limit") int limit){
        return sharingRegistryService.getUsers(domainId,offset, limit);
    }
}
