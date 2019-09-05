package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.User;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/domains/{domainId}/users")
public class UsersController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to register a user in the system</p>
     */
    @RequestMapping(method = RequestMethod.POST)
    public ResponseEntity<User> createUser(@Valid @RequestBody User user){
        User createdUser = sharingRegistryService.createUser(user);
        return new ResponseEntity<>(createdUser, HttpStatus.CREATED);
    }

    /**
     <p>API method to update existing user</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateUser(@Valid @RequestBody User user){
        sharingRegistryService.updatedUser(user);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to check User Exists</p>
     */
    @RequestMapping(value = "/{userId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isUserExists(@PathVariable("domainId") String domainId, @PathVariable("userId") String userId){
        boolean response = sharingRegistryService.isUserExists(domainId, userId);
        if(response){
            return new ResponseEntity<>(HttpStatus.OK);
        }else{
            return new ResponseEntity<>(HttpStatus.NOT_FOUND);
        }
    }

    /**
     <p>API method to delete user</p>
     */
    @RequestMapping(value="/{userId}",method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteUser(@PathVariable("userId") String userId, @PathVariable("domainId") String domainId){
        sharingRegistryService.deleteUser(domainId, userId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get a user</p>
     */
    @RequestMapping(value = "/{userId}", method = RequestMethod.GET)
    public ResponseEntity<User> getUser(@PathVariable("domainId") String domainId,@PathVariable("userId") String userId){
        User user = sharingRegistryService.getUser(domainId, userId);
        return new ResponseEntity<>(user, HttpStatus.OK);
    }

    /**
     <p>API method to get a list of users in a specific domain.</p>
     <li>domainId : Domain id</li>
     <li>offset : Starting result number</li>
     <li>limit : Number of max results to be sent</li>
     */
    @RequestMapping(method = RequestMethod.GET)
    public List<User> getUsers(@PathVariable("domainId") String domainId, @RequestParam(name = "offset", defaultValue = "0") int offset, @RequestParam(name = "limit", defaultValue = "-1") int limit){
        return sharingRegistryService.getUsers(domainId,offset, limit);
    }
}
